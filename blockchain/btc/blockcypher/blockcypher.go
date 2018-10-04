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

// Package blockcypher defines primitives to work with the BlockCypher API.
package blockcypher

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/types"
)

const (
	// DefaultLimiterInterval is the default BlockCypher API limiter
	// interval.
	DefaultLimiterInterval = time.Minute

	// DefaultLimiterSize is the default BlockCypher API limiter size.
	DefaultLimiterSize = 2
)

// Config contains configuration options for the client.
type Config struct {
	// Network is the Bitcoin network.
	Network btc.Network

	// APIKey is an optional BlockCypher API key.
	APIKey string

	// LimiterInterval is the BlockCypher API limiter interval.
	LimiterInterval time.Duration

	// LimiterSize is the BlockCypher API limiter size.
	LimiterSize int
}

// Client is a BlockCypher API client.
type Client struct {
	config    *Config
	api       *gobcy.API
	limiter   chan struct{}
	timer     *time.Timer
	waitGroup sync.WaitGroup
}

// New creates a client for a Bitcoin network, using an optional BlockCypher API
// key.
func New(c *Config) *Client {
	parts := strings.Split(c.Network.String(), ":")
	size := c.LimiterSize
	if size == 0 {
		size = DefaultLimiterSize
	}
	limiter := make(chan struct{}, size)

	return &Client{
		config:  c,
		api:     &gobcy.API{Token: c.APIKey, Coin: "btc", Chain: parts[1]},
		limiter: limiter,
	}
}

// FindUnspent implements
// github.com/stratumn/go-core/blockchain/btc.UnspentFinder.FindUnspent.
func (c *Client) FindUnspent(address *types.ReversedBytes20, amount int64) (res btc.UnspentResult, err error) {
	for range c.limiter {
		break
	}
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	addr := base58.CheckEncode(address[:], c.config.Network.ID())
	addrInfo, err := c.api.GetAddr(addr, map[string]string{
		"unspentOnly":   "true",
		"includeScript": "true",
		"limit":         "50",
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
			return
		}

		res.Outputs = append(res.Outputs, output)
		res.Sum += int64(TXRef.Value)

		if res.Sum >= amount {
			return
		}
	}

	err = fmt.Errorf(
		"not enough Bitcoins available on %s, expected at least %d satoshis got %d",
		addr,
		amount,
		res.Sum,
	)

	return
}

// Broadcast implements
// github.com/stratumn/go-core/blockchain/btc.Broadcaster.Broadcast.
func (c *Client) Broadcast(raw []byte) error {
	for range c.limiter {
		break
	}
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	_, err := c.api.PushTX(hex.EncodeToString(raw))
	return err
}

// Start starts the client.
func (c *Client) Start(ctx context.Context) {
	size := c.config.LimiterSize
	if size == 0 {
		size = DefaultLimiterSize
	}
	for i := 0; i < size; i++ {
		c.limiter <- struct{}{}
	}

	interval := c.config.LimiterInterval
	if interval == 0 {
		interval = DefaultLimiterInterval
	}

	c.timer = time.NewTimer(interval)

	for {
		select {
		case <-c.timer.C:
			c.timer = time.NewTimer(interval)
			c.limiter <- struct{}{}
		case <-ctx.Done():
			c.stop()
			return
		}

	}
}

// stop stops the client.
func (c *Client) stop() {
	if !c.timer.Stop() {
		<-c.timer.C
	}

	c.waitGroup.Wait()
	<-c.limiter
	close(c.limiter)
}
