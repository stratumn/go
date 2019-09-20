// Copyright 2016-2018 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package btctimestamper implements a fake Bitcoin timestamper which can be
// used for testing.
package btctimestamper

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-core/blockchain"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

const (
	// Component name for monitoring.
	Component = "btc"

	// Description describes this Timestamper.
	Description = "Bitcoin Timestamper"
)

// Config contains configuration options for the timestamper.
type Config struct {
	// An unspent transaction finder.
	UnspentFinder btc.UnspentFinder

	// A transaction broadcaster.
	Broadcaster btc.Broadcaster

	// A wallet import format key.
	WIF string

	// Transaction fee
	Fee int64
}

// Timestamper is the type that implements
// github.com/stratumn/go-core/blockchain.Timestamper.
type Timestamper struct {
	config           *Config
	net              btc.Network
	netParams        *chaincfg.Params
	privKey          *btcec.PrivateKey
	pubKey           *btcec.PublicKey
	address          *btcutil.AddressPubKeyHash
	compressedPubKey bool
}

// New creates an instance of a Timestamper.
func New(config *Config) (*Timestamper, error) {
	WIF, err := btcutil.DecodeWIF(config.WIF)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, Component, "could not create timestamper")
	}

	ts := &Timestamper{
		config:           config,
		privKey:          WIF.PrivKey,
		pubKey:           WIF.PrivKey.PubKey(),
		compressedPubKey: WIF.CompressPubKey,
	}

	if WIF.IsForNet(&chaincfg.TestNet3Params) {
		ts.net = btc.NetworkTest3
		ts.netParams = &chaincfg.TestNet3Params
	} else if WIF.IsForNet(&chaincfg.MainNetParams) {
		ts.net = btc.NetworkMain
		ts.netParams = &chaincfg.MainNetParams
	}

	if ts.netParams == nil {
		return nil, types.NewError(errorcode.InvalidArgument, Component, "unsupported network")
	}

	var pubKeyHash []byte
	switch WIF.CompressPubKey {
	case true:
		pubKeyHash = btcutil.Hash160(ts.pubKey.SerializeCompressed())
	case false:
		pubKeyHash = btcutil.Hash160(ts.pubKey.SerializeUncompressed())
	}

	ts.address, err = btcutil.NewAddressPubKeyHash(pubKeyHash, ts.netParams)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, Component, "could not create new address")
	}

	return ts, nil
}

// Network implements fmt.Stringer.
func (ts *Timestamper) Network() blockchain.Network {
	return ts.net
}

// GetInfo implements
// github.com/stratumn/go-core/blockchain.HashTimestamper.
func (ts *Timestamper) GetInfo() *blockchain.Info {
	return &blockchain.Info{
		Network:     ts.net,
		Description: Description,
	}
}

// TimestampHash implements
// github.com/stratumn/go-core/blockchain.HashTimestamper.
func (ts *Timestamper) TimestampHash(ctx context.Context, hash []byte) (txid types.TransactionID, err error) {
	span, ctx := monitoring.StartSpanIncomingRequest(ctx, "blockchain/btc/btctimestamper/TimestampHash")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	var prevPKScripts [][]byte

	addr := (*types.ReversedBytes20)(ts.address.Hash160())
	res, err := ts.config.UnspentFinder.FindUnspent(ctx, addr, ts.config.Fee)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	var utxoAmount int64
	for _, output := range res.Outputs {
		// DO NOT TAKE INTO ACCOUNT UTXO THAT ARE BELOW `amount`.
		// THIS IS A QUICK FIX FOR AN ISSUE THAT OCCURED IN THE FOSSILIZER :
		// when adding multiple inputs (among which some are lower than the sum of outputs) to a transaction,
		// blockcypher API returns either this error:
		// "Error validating transaction: Sum of inputs <lower_input_amount> lesser than outputs <sum_of_outputs>"
		// or the following one:
		// "Error running script for input 1 referencing <tx_hash> at 0: Script was NOT verified successfully.."
		// The solution consists in only adding as input an UTXO higher than `amount`.
		if int64(output.Value) > ts.config.Fee {
			prevPKScripts = append(prevPKScripts, output.PKScript)
			out := wire.NewOutPoint((*chainhash.Hash)(&output.TXHash), uint32(output.Index))
			tx.AddTxIn(wire.NewTxIn(out, nil, nil))
			utxoAmount = int64(output.Value)
		}
	}

	// If no UTXO were high enough to cover the transaction fee, abort the transaction
	if len(tx.TxIn) == 0 {
		return nil, types.NewErrorf(errorcode.FailedPrecondition,
			Component,
			"adress %s: no UTXO greater than transaction fee amount, refill required",
			addr,
		)
	}

	payToAddrOut, err := ts.createPayToAddrTxOut(utxoAmount - ts.config.Fee)
	if err != nil {
		return nil, err
	}
	tx.AddTxOut(payToAddrOut)

	nullDataOut, err := ts.createNullDataTxOut(hash)
	if err != nil {
		return nil, err
	}
	tx.AddTxOut(nullDataOut)

	if err = ts.signTx(tx, prevPKScripts); err != nil {
		return nil, err
	}
	if err = ts.validateTx(tx, prevPKScripts); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err = tx.Serialize(buf); err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, Component, "could not read tx buffer")
	}
	err = ts.config.Broadcaster.Broadcast(ctx, raw)
	if err != nil {
		return nil, err
	}

	// Reverse the bytes!
	var txHash32 types.Bytes32
	for i, b := range tx.TxHash() {
		txHash32[types.Bytes32Size-i-1] = b
	}

	remaining := res.Total - ts.config.Fee

	monitoring.TxLogEntry(ctx).WithFields(log.Fields{
		"txid":      txHash32,
		"remaining": remaining,
	}).Info("Broadcasted transaction")

	return txHash32[:], nil
}

func (ts *Timestamper) createPayToAddrTxOut(amount int64) (*wire.TxOut, error) {
	PKScript, err := txscript.PayToAddrScript(ts.address)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unknown, Component, "could not create pay-to-addr tx")
	}

	return wire.NewTxOut(amount, PKScript), nil
}

func (ts *Timestamper) createNullDataTxOut(hash []byte) (*wire.TxOut, error) {
	PKScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(hash).Script()
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unknown, Component, "could not create null tx")
	}

	return wire.NewTxOut(0, PKScript), nil
}

func (ts *Timestamper) signTx(tx *wire.MsgTx, prevPKScripts [][]byte) error {
	for index, PKScript := range prevPKScripts {
		sig, err := txscript.SignTxOutput(
			ts.netParams,
			tx,
			0,
			PKScript,
			txscript.SigHashAll,
			txscript.KeyClosure(ts.lookupKey),
			nil,
			nil)
		if err != nil {
			return types.WrapError(err, errorcode.InvalidArgument, Component, "could not sign tx")
		}

		tx.TxIn[index].SignatureScript = sig
	}

	return nil
}

const validateTxEngineFlags = txscript.ScriptBip16 | txscript.ScriptVerifyDERSignatures |
	txscript.ScriptStrictMultiSig | txscript.ScriptDiscourageUpgradableNops

func (ts *Timestamper) validateTx(tx *wire.MsgTx, prevPKScripts [][]byte) error {
	for _, PKScript := range prevPKScripts {
		vm, err := txscript.NewEngine(PKScript, tx, 0, validateTxEngineFlags, nil, nil, 0)
		if err != nil {
			return types.WrapError(err, errorcode.InvalidArgument, Component, "could not create tx script engine")
		}
		if err := vm.Execute(); err != nil {
			return types.WrapError(err, errorcode.InvalidArgument, Component, "tx validation failed")
		}
	}

	return nil
}

func (ts *Timestamper) lookupKey(btcutil.Address) (*btcec.PrivateKey, bool, error) {
	return ts.privKey, ts.compressedPubKey, nil
}
