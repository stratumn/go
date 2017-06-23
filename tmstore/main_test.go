package tmstore

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)

	// start a tendermint node (and tmpop app) in the background to test against
	node := StartNode()
	flag.Parse()
	code := m.Run()

	// and shut down proper at the end
	node.Stop()
	node.Wait()
	os.Exit(code)
}
