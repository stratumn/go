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

// Package blockcypher defines primitives to work with the BlockCypher API.
package blockcypher

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"

	"go.opencensus.io/trace"
)

const (
	// Component name for monitoring.
	Component = "btc"
)

// Config contains configuration options for the client.
type Config struct {
	// Network is the Bitcoin network.
	Network btc.Network

	// APIKey is an optional BlockCypher API key.
	// It prevents your requests from being throttled.
	APIKey string
}

// Client is a BlockCypher API client.
type Client struct {
	config *Config
	api    *gobcy.API
}

// New creates a client for a Bitcoin network, using an optional BlockCypher API
// key.
func New(c *Config) *Client {
	parts := strings.Split(c.Network.String(), ":")

	return &Client{
		config: c,
		api:    &gobcy.API{Token: c.APIKey, Coin: "btc", Chain: parts[1]},
	}
}

// FindUnspent implements
// github.com/stratumn/go-core/blockchain/btc.UnspentFinder.FindUnspent.
func (c *Client) FindUnspent(ctx context.Context, address *types.ReversedBytes20, amount int64) (res btc.UnspentResult, err error) {
	_, span := trace.StartSpan(ctx, "blockchain/btc/blockcypher/FindUnspent")
	defer span.End()

	addr := base58.CheckEncode(address[:], c.config.Network.ID())
	var addrInfo gobcy.Addr
	err = RetryWithBackOff(ctx, func() error {
		addrInfo, err = c.api.GetAddr(addr, map[string]string{
			"unspentOnly":   "true",
			"includeScript": "true",
			"limit":         "50",
		})

		return err
	})
	if err != nil {
		return
	}

	res.Total = int64(addrInfo.Balance)

	for _, TXRef := range addrInfo.TXRefs {
		output := btc.Output{Index: TXRef.TXOutputN}

		if err = output.TXHash.Unstring(TXRef.TXHash); err != nil {
			return
		}

		output.PKScript, err = hex.DecodeString(TXRef.Script)
		if err != nil {
			return res, types.WrapError(err, errorcode.InvalidArgument, Component, "invalid tx script")
		}

		res.Outputs = append(res.Outputs, output)
		res.Sum += int64(TXRef.Value)

		if res.Sum >= amount {
			return
		}
	}

	err = types.NewErrorf(errorcode.FailedPrecondition,
		Component,
		"not enough Bitcoins available on %s, expected at least %d satoshis got %d",
		addr,
		amount,
		res.Sum,
	)

	return
}

// Broadcast implements
// github.com/stratumn/go-core/blockchain/btc.Broadcaster.Broadcast.
func (c *Client) Broadcast(ctx context.Context, raw []byte) error {
	_, span := trace.StartSpan(ctx, "blockchain/btc/blockcypher/Broadcast")
	defer span.End()

	return RetryWithBackOff(ctx, func() error {
		_, err := c.api.PushTX(hex.EncodeToString(raw))
		return err
	})
}
