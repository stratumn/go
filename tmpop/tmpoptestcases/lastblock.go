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
	"testing"

	"github.com/stratumn/sdk/tmpop"
)

// TestLastBlock tests if tmpop correctly stores information
// about the previous block and previous history when committing.
func (f Factory) TestLastBlock(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	link1, req := commitRandomLink(t, h, req)
	link2, req := commitRandomLink(t, h, req)
	lastAppHash := req.Header.AppHash

	t.Run("Commit stores last block information", func(t *testing.T) {
		got, err := tmpop.ReadLastBlock(f.kv)
		if err != nil {
			t.Fatal(err)
		}

		if got.Height != 2 {
			t.Errorf("a.Commit(): expected commit to save the last block height, got %v, expected %v",
				got.Height, 2)
		}

		if !got.AppHash.EqualsBytes(lastAppHash) {
			t.Errorf("a.Commit(): expected commit to save the last app hash, expected %v, got %v",
				lastAppHash, got.AppHash)
		}
	})

	t.Run("Restart with existing history", func(t *testing.T) {
		h2, err := tmpop.New(f.adapter, f.kv, &tmpop.Config{})
		if err != nil {
			t.Fatalf("Couldn't start tmpop with existing store: %s", err)
		}

		verifyLinkStored(t, h2, link1)
		verifyLinkStored(t, h2, link2)
	})
}
