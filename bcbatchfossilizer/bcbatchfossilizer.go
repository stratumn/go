// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// Package bcbatchfossilizer implements a fossilizer that fossilize batches of
// hashes on a blockchain.
package bcbatchfossilizer

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/evidences"
	"github.com/stratumn/sdk/fossilizer"
	"github.com/stratumn/sdk/types"

	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/blockchain"
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

// Fossilizer is the type that
// implements github.com/stratumn/sdk/fossilizer.Adapter.
type Fossilizer struct {
	*batchfossilizer.Fossilizer
	config            *Config
	lastRoot          *types.Bytes32
	lastTransactionID types.TransactionID
}

// New creates an instance of a Fossilizer.
func New(config *Config, batchConfig *batchfossilizer.Config) (*Fossilizer, error) {
	if batchConfig.MaxSimBatches > 1 {
		return nil, fmt.Errorf("MaxSimBatches is %d want less than 2", batchConfig.MaxSimBatches)
	}

	b, err := batchfossilizer.New(batchConfig)
	if err != nil {
		return nil, err
	}

	f := Fossilizer{
		Fossilizer: b,
		config:     config,
	}

	f.SetTransformer(f.transform)

	return &f, err
}

// GetInfo implements github.com/stratumn/sdk/fossilizer.Adapter.GetInfo.
func (a *Fossilizer) GetInfo() (interface{}, error) {
	batchInfo, err := a.Fossilizer.GetInfo()
	if err != nil {
		return nil, err
	}

	info, ok := batchInfo.(*batchfossilizer.Info)
	if !ok {
		return nil, fmt.Errorf("unexpected batchfossilizer info %#v", batchInfo)
	}

	return &Info{
		Name:        Name,
		Description: Description,
		Version:     info.Version,
		Commit:      info.Commit,
		Blockchain:  a.config.HashTimestamper.Network().String(),
	}, nil
}

func (a *Fossilizer) transform(evidence *cs.Evidence, data, meta []byte) (*fossilizer.Result, error) {
	var (
		root = evidence.Proof.(*evidences.BatchProof).Root
		txid types.TransactionID
		err  error
	)

	if a.lastRoot == nil || *root != *a.lastRoot {
		txid, err = a.config.HashTimestamper.TimestampHash(root)
		if err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"txid": txid,
			"root": root,
		}).Info("Broadcasted transaction")

		a.lastRoot = root
		a.lastTransactionID = txid
	}

	evidence.Provider = a.config.HashTimestamper.Network().String()
	evidence.Backend = Name
	evidence.Proof = &evidences.BcBatchProof{
		Batch:         *evidence.Proof.(*evidences.BatchProof),
		TransactionID: a.lastTransactionID,
	}

	r := fossilizer.Result{
		Evidence: *evidence,
		Data:     data,
		Meta:     meta,
	}

	return &r, nil
}
