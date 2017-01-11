package tmstore

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// start a tendermint node (and tmpop app) in the background to test against
	StartNode()
	os.Exit(m.Run())
}
