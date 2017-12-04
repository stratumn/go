package tmstore

/**
This file is base HEAVILY on tendermint/tendermint/rpc/tests/helpers.go
However, I wanted to use public variables, so this could be a basis
of tests in various packages.
**/

import (
	"path/filepath"

	"github.com/stratumn/sdk/dummystore"
	"github.com/stratumn/sdk/tmpop"
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
	return NewFromClient(&Config{}, func(endpoint string) client.Client {
		return client.NewLocal(testNode)
	})
}

func ResetNode() {
	testNode.Reset()
	*testDummyStore = *dummystore.New(&dummystore.Config{})
}

func StartNode() *node.Node {
	testDummyStore = dummystore.New(&dummystore.Config{})
	var err error
	testTmpop, err = tmpop.New(testDummyStore, testDummyStore, &tmpop.Config{
		ValidatorFilename: filepath.Join("testdata", "rules.json"),
	})
	if err != nil {
		panic(err)
	}

	testNode = rpctest.StartTendermint(testTmpop)
	testClient := tmpop.NewTendermintClient(client.NewLocal(testNode))
	testTmpop.ConnectTendermint(testClient)

	return testNode
}
