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
	"context"
	"encoding/json"
	"fmt"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
)

// GetCurrentHeader returns the current block header
func (t *TMPop) GetCurrentHeader() *abci.Header {
	return t.currentHeader
}

// ReadLastBlock returns the last block committed by TMPop
func ReadLastBlock(ctx context.Context, kv store.KeyValueReader) (*LastBlock, error) {
	lBytes, err := kv.GetValue(ctx, tmpopLastBlockKey)
	if err != nil {
		return nil, err
	}

	var l LastBlock
	if lBytes == nil {
		return &l, nil
	}
	err = wire.ReadBinaryBytes(lBytes, &l)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

// saveLastBlock saves the last block committed by TMPop
func saveLastBlock(ctx context.Context, a store.KeyValueWriter, l LastBlock) {
	a.SetValue(ctx, tmpopLastBlockKey, wire.BinaryBytes(l))
}

func getValidatorHashKey(height int64) []byte {
	key := fmt.Sprintf("tmpop:validator:%d", height)
	return []byte(key)
}

// saveValidatorHash saves the hash of the validator used for the current block
func (t *TMPop) saveValidatorHash(ctx context.Context) error {
	if t.state.validator != nil {
		key := getValidatorHashKey(t.currentHeader.Height)
		h, err := t.state.validator.Hash()
		if err != nil {
			return err
		}

		if err := t.kvDB.SetValue(ctx, key, h[:]); err != nil {
			return err
		}
	}

	return nil
}

// getValidatorHash gets the hash of the validator used for a block at a specific height
func (t *TMPop) getValidatorHash(ctx context.Context, height int64) (*types.Bytes32, error) {
	key := getValidatorHashKey(height)
	value, err := t.kvDB.GetValue(ctx, key)
	if err != nil || value == nil {
		return nil, err
	}

	return types.NewBytes32FromBytes(value), nil
}

func getCommitLinkHashesKey(height int64) []byte {
	key := fmt.Sprintf("tmpop:linkhashes:%d", height)
	return []byte(key)
}

// saveCommitLinkHashes saves the link hashes of valid links created in the
// current block. Since Tendermint can include invalid transactions and
// doesn't provide yet an easy way to know which transactions are invalid in
// a block, this is useful to generate valid evidence and ignore invalid
// transactions.
func (t *TMPop) saveCommitLinkHashes(ctx context.Context, links []*cs.Link) error {
	if len(links) > 0 {
		key := getCommitLinkHashesKey(t.currentHeader.Height)

		var linkHashes []types.Bytes32
		for _, link := range links {
			linkHash, _ := link.Hash()
			linkHashes = append(linkHashes, *linkHash)
		}

		value, err := json.Marshal(linkHashes)
		if err != nil {
			return err
		}

		if err := t.kvDB.SetValue(ctx, key, value); err != nil {
			return err
		}
	}

	return nil
}

// getCommitLinkHashes gets the link hashes included in a block at a specific height.
// This is useful to ignore invalid links included in that block.
func (t *TMPop) getCommitLinkHashes(ctx context.Context, height int64) ([]types.Bytes32, error) {
	key := getCommitLinkHashesKey(height)
	value, err := t.kvDB.GetValue(ctx, key)
	if err != nil || value == nil {
		return nil, err
	}

	var linkHashes []types.Bytes32
	if err := json.Unmarshal(value, &linkHashes); err != nil {
		return nil, err
	}

	return linkHashes, nil
}
