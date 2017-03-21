package tmstore

/**
This file is base HEAVILY on tendermint/tendermint/rpc/tests/helpers.go
However, I wanted to use public variables, so this could be a basis
of tests in various packages.
**/

import (
	logger "github.com/tendermint/go-logger"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/dummystore"
	"github.com/stratumn/sdk/tendermint"
	"github.com/stratumn/sdk/tmpop"
	cfg "github.com/tendermint/go-config"
	"github.com/tendermint/tendermint/config/tendermint_test"
)

var (
	config        cfg.Config
	TestSegment   = cstesting.RandomSegment()
	ToSaveSegment = cstesting.RandomSegment()
	SegmentSaved  = false
	TestLimit     = 1
	testTmpop     *tmpop.TMPop
)

const tmLogLevel = "error"

// GetConfig returns a config for the test cases as a singleton
func GetConfig() cfg.Config {
	if config == nil {
		config = tendermint_test.ResetConfig("rpc_test_client_test")
		// Shut up the logging
		logger.SetLogLevel(tmLogLevel)
	}
	return config
}

// GetClient gets a rpc client pointing to the test node
func GetClient() *TMClient {
	rpcAddr := GetConfig().GetString("rpc_laddr")
	return NewTMClient(rpcAddr)
}

// StartNode starts a test node in a go routine and returns when it is initialized
func StartNode() {
	// start a node
	ready := make(chan struct{})
	go NewNode(ready)
	<-ready
}

// NewNode creates a new node and sleeps forever
func NewNode(ready chan struct{}) {
	adapter := dummystore.New(&dummystore.Config{})
	var err error
	testTmpop, err = tmpop.New(adapter, &tmpop.Config{})
	if err != nil {
		panic(err)
	}

	tendermint.RunNode(GetConfig(), testTmpop)

	ready <- struct{}{}

	// Sleep forever
	ch := make(chan struct{})
	<-ch
}

func Reset() {
	a := dummystore.New(&dummystore.Config{})
	testTmpop.ResetAdapter(a)
}
