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

//go:generate mockgen -package tmpoptestcasesmocks -destination tmpoptestcases/mocks/mocktmclient.go github.com/stratumn/go-core/tmpop TendermintClient

package tmpop

import (
	"bytes"
	"context"
	"fmt"

	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/tmpop/evidences"
	"github.com/stratumn/go-core/types"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"go.elastic.co/apm"
)

// TendermintClient is a light interface to query Tendermint Core.
type TendermintClient interface {
	Block(ctx context.Context, height int64) (*Block, error)
}

// Block contains the parts of a Tendermint block that TMPoP is interested in.
type Block struct {
	// The block's header.
	Header *tmtypes.Header
	// A block at height N contains the votes for block N-1.
	Votes []*evidences.TendermintVote
	// A block at height N contains the validator set for block N-1.
	Validators *tmtypes.ValidatorSet
	// The block's transactions.
	Txs []*Tx
}

// TendermintClientWrapper implements TendermintClient.
type TendermintClientWrapper struct {
	tmClient client.Client
}

// NewTendermintClient creates a new TendermintClient.
func NewTendermintClient(tmClient client.Client) *TendermintClientWrapper {
	return &TendermintClientWrapper{
		tmClient: tmClient,
	}
}

// Block queries for a block at a specific height.
func (c *TendermintClientWrapper) Block(ctx context.Context, height int64) (*Block, error) {
	span, _ := apm.StartSpan(ctx, "tmclient/Block", monitoring.SpanTypeIncomingRequest)
	defer span.End()

	tmBlock, err := c.tmClient.Block(&height)
	if err != nil {
		err = types.WrapError(err, errorcode.Unavailable, Name, "could not get block from Tendermint Core")
		monitoring.SetSpanStatus(span, err)
		return nil, err
	}

	// The votes in block N are voting on block N-1.
	// So we need the validator set of the previous block,
	// unless it's the genesis block.
	prevHeight := height - 1
	if prevHeight <= 0 {
		prevHeight = 1
	}
	validators, err := c.tmClient.Validators(&prevHeight)
	if err != nil {
		err = types.WrapError(err, errorcode.Unavailable, Name, "could not get validators from Tendermint Core")
		monitoring.SetSpanStatus(span, err)
		return nil, err
	}

	block := &Block{
		Header:     tmBlock.BlockMeta.Header,
		Validators: &tmtypes.ValidatorSet{Validators: validators.Validators},
	}

	for _, v := range tmBlock.Block.LastCommit.Precommits {
		// If a block is invalid, non-byzantine validators
		// will send a nil vote.
		if v == nil {
			continue
		}

		vote, err := getVote(v, validators)
		if err != nil {
			return nil, err
		}

		block.Votes = append(block.Votes, vote)
	}

	for _, tx := range tmBlock.Block.Txs {
		tmTx, err := unmarshallTx(tx)
		if !err.IsOK() || tmTx.TxType != CreateLink {
			monitoring.LogWithTxFields(ctx).Warnf("Could not unmarshall block Tx %+v. Evidence will not be created.", tx)
			span.Context.SetTag(monitoring.ErrorLabel, fmt.Sprintf("Could not unmarshall block Tx %+v.", tx))
			continue
		}

		block.Txs = append(block.Txs, tmTx)
	}

	return block, nil
}

func getVote(v *tmtypes.Vote, validators *ctypes.ResultValidators) (*evidences.TendermintVote, error) {
	for _, val := range validators.Validators {
		if bytes.Equal(v.ValidatorAddress.Bytes(), val.Address.Bytes()) {
			return &evidences.TendermintVote{PubKey: &val.PubKey, Vote: v}, nil
		}
	}

	return nil, types.NewErrorf(errorcode.InvalidArgument, Name, "could not find validator address %x", v.ValidatorAddress)
}
