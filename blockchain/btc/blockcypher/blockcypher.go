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
	"time"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
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
	span, ctx := monitoring.StartSpanOutgoingRequest(ctx, "blockchain/btc/blockcypher/FindUnspent")
	start := time.Now()
	defer func() {
		if err != nil {
			requestErr.With(prometheus.Labels{requestType: "FindUnspent"}).Inc()
		}

		requestCount.With(prometheus.Labels{requestType: "FindUnspent"}).Inc()
		requestLatency.With(prometheus.Labels{requestType: "FindUnspent"}).Observe(
			float64(time.Since(start)) / float64(time.Millisecond),
		)

		monitoring.SetSpanStatusAndEnd(span, err)
	}()

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
	accountBalance.With(prometheus.Labels{accountAddress: addr}).Set(float64(res.Total))

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
	span, ctx := monitoring.StartSpanOutgoingRequest(ctx, "blockchain/btc/blockcypher/Broadcast")
	defer span.End()

	start := time.Now()
	err := RetryWithBackOff(ctx, func() error {
		_, err := c.api.PushTX(hex.EncodeToString(raw))
		return err
	})
	if err != nil {
		requestErr.With(prometheus.Labels{requestType: "Broadcast"}).Inc()
		monitoring.SetSpanStatus(span, err)
	}

	requestCount.With(prometheus.Labels{requestType: "Broadcast"}).Inc()
	requestLatency.With(prometheus.Labels{requestType: "Broadcast"}).Observe(
		float64(time.Since(start)) / float64(time.Millisecond),
	)

	return err
}
