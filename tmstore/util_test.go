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

package tmstore

/**
This file is base HEAVILY on tendermint/tendermint/rpc/tests/helpers.go
However, I wanted to use public variables, so this could be a basis
of tests in various packages.
**/

import (
	"context"

	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/tmpop"
	node "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/client"
	rpctest "github.com/tendermint/tendermint/rpc/test"
)

var (
	testDummyStore *dummystore.DummyStore
	testTmpop      *tmpop.TMPop
	testNode       *node.Node
)

// NewTestClient returns a rpc client pointing to the test node
func NewTestClient() *TMStore {
	return New(&Config{}, client.NewLocal(testNode))
}

func ResetNode() {
	testNode.Reset()
	*testDummyStore = *dummystore.New(&dummystore.Config{})
}

func StartNode(config *tmpop.Config) *node.Node {
	testDummyStore = dummystore.New(&dummystore.Config{})
	var err error
	testTmpop, err = tmpop.New(context.Background(), testDummyStore, testDummyStore, config)
	if err != nil {
		panic(err)
	}

	testNode = rpctest.NewTendermint(testTmpop)
	testClient := tmpop.NewTendermintClient(client.NewLocal(testNode))
	testTmpop.ConnectTendermint(testClient)

	if err = testNode.Start(); err != nil {
		panic(err)
	}

	return testNode
}
