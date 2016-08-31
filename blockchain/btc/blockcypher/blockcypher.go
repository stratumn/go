// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// Package blockcypher defines primitives to work with the BlockCypher API.
package blockcypher

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcutil/base58"
	"github.com/stratumn/go/types"
	"github.com/stratumn/goprivate/blockchain/btc"
)

const (
	// DefaultLimiterInterval is the default BlockCypher API limiter interval.
	DefaultLimiterInterval = 2 * time.Minute

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
	closeChan chan struct{}
	timer     *time.Timer
	waitGroup sync.WaitGroup
}

// New creates a client for a Bitcoin network, using an optional BlockCypher API key.
func New(c *Config) *Client {
	parts := strings.Split(c.Network.String(), ":")
	size := c.LimiterSize
	if size == 0 {
		size = DefaultLimiterSize
	}
	limiter := make(chan struct{}, size)

	return &Client{
		config:    c,
		api:       &gobcy.API{c.APIKey, "btc", parts[1]},
		limiter:   limiter,
		closeChan: make(chan struct{}),
	}
}

// FindUnspent implements github.com/stratumn/goprivate/blockchain/btc.UnspentFinder.FindUnspent.
func (c *Client) FindUnspent(address *types.ReversedBytes20, amount int64) ([]btc.Output, int64, error) {
	if err := c.wait(); err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}

	var (
		outputs []btc.Output
		total   int64
	)

TX_LOOP:
	for _, TXRef := range addrInfo.TXRefs {
		output := btc.Output{Index: TXRef.TXOutputN}
		if err := output.TXHash.Unstring(TXRef.TXHash); err != nil {
			return nil, 0, err
		}

		output.PKScript, err = hex.DecodeString(TXRef.Script)
		if err != nil {
			return nil, 0, err
		}

		outputs = append(outputs, output)

		total += int64(TXRef.Value)
		if total >= amount {
			break TX_LOOP
		}
	}

	if total < amount {
		return nil, 0, fmt.Errorf("could not get amount %d got %d", amount, total)
	}

	return outputs, total, nil
}

// Broadcast implements github.com/stratumn/goprivate/blockchain/btc.Broadcaster.Broadcast.
func (c *Client) Broadcast(raw []byte) error {
	if err := c.wait(); err != nil {
		return err
	}
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	_, err := c.api.PushTX(hex.EncodeToString(raw))
	return err
}

// Start starts the client.
func (c *Client) Start() {
	size := c.config.LimiterSize
	if size == 0 {
		size = DefaultLimiterSize
	}
	interval := c.config.LimiterInterval
	if interval == 0 {
		interval = DefaultLimiterInterval
	}
	for i := 0; i < size; i++ {
		c.limiter <- struct{}{}
	}

	c.timer = time.NewTimer(interval)

	go func() {
		for {
			for _ = range c.timer.C {
				break
			}
			if c.closeChan == nil {
				return
			}
			c.timer = time.NewTimer(interval)
			c.limiter <- struct{}{}
		}
	}()

	<-c.closeChan
}

// Stop stops the client.
func (c *Client) Stop() {
	c.closeChan <- struct{}{}
	close(c.closeChan)
	c.closeChan = nil

	if !c.timer.Stop() {
		<-c.timer.C
	}

	c.waitGroup.Wait()
	<-c.limiter
	close(c.limiter)
}

func (c *Client) wait() error {
	for _ = range c.limiter {
		break
	}
	if c.closeChan == nil {
		return errors.New("client is stopped")
	}
	return nil
}
