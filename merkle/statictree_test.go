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

package merkle_test

import (
	"testing"

	"github.com/stratumn/sdk/types"
	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/merkle/treetestcases"
)

func TestNewStaticTree_noLeaves(t *testing.T) {
	_, err := merkle.NewStaticTree(nil)
	if err == nil {
		t.Error("NewStaticTree(): err = nil want Error")
	}
	if got, want := err.Error(), "tree should have at least one leaf"; got != want {
		t.Errorf("NewStaticTree(): err.Error() = %q want %q", got, want)
	}
}

func TestStaticTree(t *testing.T) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			return merkle.NewStaticTree(leaves)
		},
	}.RunTests(t)
}

func BenchmarkStaticTree(b *testing.B) {
	treetestcases.Factory{
		New: func(leaves []types.Bytes32) (merkle.Tree, error) {
			return merkle.NewStaticTree(leaves)
		},
	}.RunBenchmarks(b)
}
