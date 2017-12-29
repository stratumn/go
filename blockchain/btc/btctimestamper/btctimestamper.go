// Copyright 2017 Stratumn SAS. All rights reserved.
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
	"errors"
	"io/ioutil"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/stratumn/sdk/blockchain"
	"github.com/stratumn/sdk/blockchain/btc"
	"github.com/stratumn/sdk/types"
)

const (
	// DefaultFee is the default transaction fee.
	DefaultFee = int64(15000)

	// Description describes this Timestamper
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
// github.com/stratumn/sdk/blockchain.Timestamper.
type Timestamper struct {
	config    *Config
	net       btc.Network
	netParams *chaincfg.Params
	privKey   *btcec.PrivateKey
	pubKey    *btcec.PublicKey
	address   *btcutil.AddressPubKeyHash
}

// New creates an instance of a Timestamper.
func New(config *Config) (*Timestamper, error) {
	WIF, err := btcutil.DecodeWIF(config.WIF)
	if err != nil {
		return nil, err
	}

	ts := &Timestamper{
		config:  config,
		privKey: WIF.PrivKey,
		pubKey:  WIF.PrivKey.PubKey(),
	}

	if WIF.IsForNet(&chaincfg.TestNet3Params) {
		ts.net = btc.NetworkTest3
		ts.netParams = &chaincfg.TestNet3Params
	} else if WIF.IsForNet(&chaincfg.MainNetParams) {
		ts.net = btc.NetworkMain
		ts.netParams = &chaincfg.MainNetParams
	}

	if ts.netParams == nil {
		return nil, errors.New("unsupported network")
	}

	pubKeyHash := btcutil.Hash160(ts.pubKey.SerializeUncompressed())
	ts.address, err = btcutil.NewAddressPubKeyHash(pubKeyHash, ts.netParams)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// Network implements fmt.Stringer.
func (ts *Timestamper) Network() blockchain.Network {
	return ts.net
}

// GetInfo implements
// github.com/stratumn/sdk/blockchain.HashTimestamper.
func (ts *Timestamper) GetInfo() *blockchain.Info {
	return &blockchain.Info{
		Network:     ts.net,
		Description: Description,
	}
}

// TimestampHash implements
// github.com/stratumn/sdk/blockchain.HashTimestamper.
func (ts *Timestamper) TimestampHash(hash *types.Bytes32) (types.TransactionID, error) {
	var prevPKScripts [][]byte

	addr := (*types.ReversedBytes20)(ts.address.Hash160())
	outputs, total, err := ts.config.UnspentFinder.FindUnspent(addr, ts.config.Fee)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	for _, output := range outputs {
		prevPKScripts = append(prevPKScripts, output.PKScript)
		out := wire.NewOutPoint((*chainhash.Hash)(&output.TXHash), uint32(output.Index))
		tx.AddTxIn(wire.NewTxIn(out, nil, nil))
	}

	payToAddrOut, err := ts.createPayToAddrTxOut(total - ts.config.Fee)
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
		return nil, err
	}
	err = ts.config.Broadcaster.Broadcast(raw)
	if err != nil {
		return nil, err
	}

	// Reverse the bytes!
	var txHash32 types.Bytes32
	for i, b := range tx.TxHash() {
		txHash32[types.Bytes32Size-i-1] = b
	}

	return txHash32[:], nil
}

func (ts *Timestamper) createPayToAddrTxOut(amount int64) (*wire.TxOut, error) {
	PKScript, err := txscript.PayToAddrScript(ts.address)
	if err != nil {
		return nil, err
	}

	return wire.NewTxOut(amount, PKScript), nil
}

func (ts *Timestamper) createNullDataTxOut(hash *types.Bytes32) (*wire.TxOut, error) {
	PKScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddData(hash[:]).Script()
	if err != nil {
		return nil, err
	}

	return wire.NewTxOut(0, PKScript), nil
}

func (ts *Timestamper) signTx(tx *wire.MsgTx, prevPKScripts [][]byte) error {
	for index, PKScript := range prevPKScripts {
		sig, err := txscript.SignTxOutput(ts.netParams, tx, 0, PKScript,
			txscript.SigHashAll, txscript.KeyClosure(ts.lookupKey), nil, nil)
		if err != nil {
			return err
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
			return err
		}
		if err := vm.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func (ts *Timestamper) lookupKey(btcutil.Address) (*btcec.PrivateKey, bool, error) {
	// Second value means uncompressed.
	return ts.privKey, false, nil
}
