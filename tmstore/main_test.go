package tmstore

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// start a tendermint node (and tmpop app) in the background to test against
	node := StartNode()
	code := m.Run()

	// and shut down proper at the end
	node.Stop()
	node.Wait()
	os.Exit(code)
}
