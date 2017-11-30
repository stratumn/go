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

// Package merkle contains types and functions to create and work with Merkle
// trees.
package merkle

import (
	"github.com/stratumn/sdk/types"
)

const (
	// HashByteSize is the length of a hash or leaf measured in bytes.
	HashByteSize = types.Bytes32Size
)

// Tree must be implemented by Merkle tree implementations.
type Tree interface {
	// NumLeaves returns the number of leaves.
	LeavesLen() int

	// Leaf returns the Merkle root.
	Root() *types.Bytes32

	// Leaf returns the leaf at the specified index.
	Leaf(index int) *types.Bytes32

	// Path returns the path of a leaf to the Merkle root.
	Path(index int) types.Path
}
