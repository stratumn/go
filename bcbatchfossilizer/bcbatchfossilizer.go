// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// Package bcbatchfossilizer implements a fossilizer that fossilize batches of hashes on a blockchain.
package bcbatchfossilizer

import (
	"errors"
	"log"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/blockchain"
	"github.com/stratumn/goprivate/types"
)

const (
	// Name is the name set in the fossilizer's information.
	Name = "bcbatch"

	// Description is the description set in the fossilizer's information.
	Description = "Stratumn Blockchain Batch Fossilizer"
)

// Config contains configuration options for the fossilizer.
type Config struct {
	HashTimestamper blockchain.HashTimestamper
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	Blockchain  string `json:"blockchain"`
}

// Evidence is the evidence sent to the result channel.
type Evidence struct {
	*batchfossilizer.Evidence
	TransactionID blockchain.TransactionID `json:"txid"`
}

// Fossilizer is the type that implements github.com/stratumn/go/fossilizer.Adapter.
type Fossilizer struct {
	*batchfossilizer.Fossilizer
	config      *Config
	resultChans []chan *fossilizer.Result
	resultChan  chan *fossilizer.Result
}

// New creates an instance of a Fossilizer.
func New(config *Config, batchConfig *batchfossilizer.Config) (*Fossilizer, error) {
	b, err := batchfossilizer.New(batchConfig)
	if err != nil {
		return nil, err
	}

	return &Fossilizer{
		Fossilizer: b,
		config:     config,
	}, err
}

// GetInfo implements github.com/stratumn/go/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo() (interface{}, error) {
	batchInfo, err := a.Fossilizer.GetInfo()
	if err != nil {
		return nil, err
	}

	info, ok := batchInfo.(*batchfossilizer.Info)
	if !ok {
		return nil, errors.New("unexpected batchfossilizer info")
	}

	return &Info{
		Name:        Name,
		Description: Description,
		Version:     info.Version,
		Commit:      info.Commit,
		Blockchain:  a.config.HashTimestamper.Network().String(),
	}, nil
}

// AddResultChan implements github.com/stratumn/go/fossilizer.Adapter.AddResultChan.
func (a *Fossilizer) AddResultChan(resultChan chan *fossilizer.Result) {
	a.resultChans = append(a.resultChans, resultChan)
}

// Start starts the fossilizer.
func (a *Fossilizer) Start() error {
	a.resultChan = make(chan *fossilizer.Result)
	a.Fossilizer.AddResultChan(a.resultChan)

	go func() {
		var (
			err               error
			lastRoot          *types.Bytes32
			lastTransactionID blockchain.TransactionID
		)

		for r := range a.resultChan {
			batchEvidenceWrapper, ok := r.Evidence.(*batchfossilizer.EvidenceWrapper)
			if !ok {
				log.Printf("unexpected batchfossilizer evidence: %v\n", batchEvidenceWrapper)
				continue
			}

			root := &batchEvidenceWrapper.Evidence.Root

			if lastRoot == nil || *root != *lastRoot {
				lastTransactionID, err = a.config.HashTimestamper.TimestampHash(root)
				if err != nil {
					log.Println(err)
					continue
				}
			}

			evidenceWrapper := map[string]*Evidence{}
			evidenceWrapper[a.config.HashTimestamper.Network().String()] = &Evidence{
				Evidence:      batchEvidenceWrapper.Evidence,
				TransactionID: lastTransactionID,
			}
			r.Evidence = evidenceWrapper

			for _, c := range a.resultChans {
				c <- r
			}

			lastRoot = root
		}
	}()

	return a.Fossilizer.Start()
}

// Stop stops the fossilizer.
func (a *Fossilizer) Stop() {
	a.Fossilizer.Stop()
	close(a.resultChan)
}