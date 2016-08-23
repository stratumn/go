// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// Package blockcypher defines primitives to work with the BlockCypher API.
package blockcypher

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
	"github.com/stratumn/goprivate/blockchain/btc"
	"github.com/stratumn/goprivate/types"
)

// Client is a BlockCypher API client.
type Client struct {
	network btc.Network
	api     *gobcy.API
}

// New creates a client for a Bitcoin network, using an optional BlockCypher API key.
func New(network btc.Network, apiKey string) *Client {
	parts := strings.Split(network.String(), ":")

	return &Client{
		network: network,
		api:     &gobcy.API{apiKey, "btc", parts[1]},
	}
}

// FindUnspent implements github.com/stratumn/goprivate/blockchain/btc.UnspentFinder.FindUnspent.
func (c *Client) FindUnspent(address160 *types.Bytes20, amount int64) ([]btc.Output, int64, error) {
	addr := base58.CheckEncode(address160[:], c.network.ID())

	addrInfo, err := c.api.GetAddrFull(addr, map[string]string{
		"unspentOnly":         "true",
		"includeHex":          "true",
		"includeConfidence":   "false",
		"omitWalletAddresses": "true",
	})
	if err != nil {
		return nil, 0, err
	}

	var (
		outputs []btc.Output
		total   int64
	)

TX_LOOP:
	for _, TX := range addrInfo.TXs {
		for index, TXOutput := range TX.Outputs {
			output := btc.Output{Index: index}

			output.TX, err = hex.DecodeString(TX.Hex)
			if err != nil {
				return nil, 0, err
			}

			TXHash, err := hex.DecodeString(TX.Hash)
			if err != nil {
				return nil, 0, err
			}

			// Reverse the bytes!
			for i, b := range TXHash {
				output.TXHash[types.Bytes32Size-i-1] = b
			}

			outputs = append(outputs, output)
			total += int64(TXOutput.Value)
			if total >= amount {
				break TX_LOOP
			}
		}
	}

	if total < amount {
		return nil, 0, fmt.Errorf("could not get amount requested, got %d, want %d", total, amount)
	}

	return outputs, total, nil
}

// Broadcast implements github.com/stratumn/goprivate/blockchain/btc.Broadcaster.Broadcast.
func (c *Client) Broadcast(raw []byte) error {
	_, err := c.api.PushTX(hex.EncodeToString(raw))
	return err
}
