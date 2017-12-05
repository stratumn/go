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
	"github.com/tendermint/tendermint/rpc/client"
	events "github.com/tendermint/tmlibs/events"
)

// TendermintClient is a light interface to query Tendermint Core
type TendermintClient interface {
	FireEvent(event string, data events.EventData)
	Block(height int) *Block
}

// Block contains the parts of a Tendermint block that TMPoP is interested in.
type Block struct {
	Txs []*Tx
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

// FireEvent fires an event through Tendermint Core
func (c *TendermintClientWrapper) FireEvent(event string, data events.EventData) {
	c.tmClient.FireEvent(event, data)
}

// Block queries for a block at a specific height
func (c *TendermintClientWrapper) Block(height int) *Block {
	previousBlock, err := c.tmClient.Block(&height)
	if err != nil {
		log.Warnf("Could not get previous block from Tendermint Core.\nSome evidence will be missing.\nError: %v", err)
	}

	block := &Block{}
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
