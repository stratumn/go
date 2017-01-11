package tmstore

/**
This file is base HEAVILY on tendermint/tendermint/rpc/tests/helpers.go
However, I wanted to use public variables, so this could be a basis
of tests in various packages.
**/

import (
	"io/ioutil"

	logger "github.com/tendermint/go-logger"

	"github.com/stratumn/go/cs/cstesting"
	"github.com/stratumn/go/dummystore"
	"github.com/stratumn/go/tmpop"
	cfg "github.com/tendermint/go-config"
	p2p "github.com/tendermint/go-p2p"
	"github.com/tendermint/tendermint/config/tendermint_test"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmtypes "github.com/tendermint/tendermint/types"
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
// TODO: can one pass an Application in????
func StartNode() {
	// start a node
	ready := make(chan struct{})
	go NewNode(ready)
	<-ready
}

// NewNode creates a new node and sleeps forever
func NewNode(ready chan struct{}) {
	adapter := dummystore.New(&dummystore.Config{})
	config := GetConfig()
	// Get PrivValidator
	privValidatorFile := config.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	dir, _ := ioutil.TempDir("", "db")
	testTmpop = tmpop.New(adapter, &tmpop.Config{DbDir: dir})
	node := nm.NewNode(config, privValidator, proxy.NewLocalClientCreator(testTmpop))

	protocol, address := nm.ProtocolAndAddress(config.GetString("node_laddr"))
	l := p2p.NewDefaultListener(protocol, address, true)
	node.AddListener(l)
	node.Start()

	// Run the RPC server.
	node.StartRPC()
	ready <- struct{}{}

	// Sleep forever
	ch := make(chan struct{})
	<-ch
}

func Reset() {
	a := dummystore.New(&dummystore.Config{})
	testTmpop.SetAdapter(a)
}
