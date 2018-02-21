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

package tmpop

import (
	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
)

// TendermintClient is a light interface to query Tendermint Core
type TendermintClient interface {
	Block(height int) *Block
}

// Block contains the parts of a Tendermint block that TMPoP is interested in.
type Block struct {
	Header *abci.Header
	Txs    []*Tx
}

// TendermintClientWrapper implements TendermintClient
type TendermintClientWrapper struct {
	tmClient client.Client
}

// NewTendermintClient creates a new TendermintClient
func NewTendermintClient(tmClient client.Client) *TendermintClientWrapper {
	return &TendermintClientWrapper{
		tmClient: tmClient,
	}
}

// Block queries for a block at a specific height
func (c *TendermintClientWrapper) Block(height int) *Block {
	requestHeight := int64(height)
	previousBlock, err := c.tmClient.Block(&requestHeight)
	if err != nil {
		log.Warnf("Could not get previous block from Tendermint Core.\nSome evidence will be missing.\nError: %v", err)
		return &Block{}
	}

	block := &Block{
		Header: &abci.Header{
			ChainID:        previousBlock.BlockMeta.Header.ChainID,
			Height:         int64(previousBlock.BlockMeta.Header.Height),
			Time:           int64(previousBlock.BlockMeta.Header.Time.Unix()),
			LastCommitHash: previousBlock.BlockMeta.Header.LastCommitHash,
			DataHash:       previousBlock.BlockMeta.Header.DataHash,
			AppHash:        previousBlock.BlockMeta.Header.AppHash,
		},
	}

	for _, tx := range previousBlock.Block.Txs {
		tmTx, err := unmarshallTx(tx)
		if !err.IsOK() || tmTx.TxType != CreateLink {
			log.Warn("Could not unmarshall previous block Tx. Evidence will not be created.")
			continue
		}

		block.Txs = append(block.Txs, tmTx)
	}

	return block
}
