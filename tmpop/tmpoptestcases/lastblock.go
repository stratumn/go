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

package tmpoptestcases

import (
	"bytes"
	"testing"

	"github.com/stratumn/sdk/tmpop"
)

// TestBeginBlockSavesLastBlockInfo tests if tmpop correctly stored informations about the previous block
func (f Factory) TestBeginBlockSavesLastBlockInfo(t *testing.T) {
	h := f.initTMPop(t, nil)

	height := uint64(2)

	req := requestBeginBlock
	req.Header.Height = height
	hash := req.GetHeader().GetAppHash()

	h.BeginBlock(req)

	got, err := tmpop.ReadLastBlock(f.adapter)
	if err != nil {
		t.Fatal(err)
	}

	if got.Height != (height - 1) {
		t.Errorf("a.Commit(): expected BeginBlock to save the last block height, got %v, expected %v",
			got.Height, height-1)
	}

	if bytes.Compare(got.AppHash, hash) != 0 {
		t.Errorf("a.Commit(): expected BeginBlock to save the last app hash, expected %v, got %v", hash, got.AppHash)
	}
}
