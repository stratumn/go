package btctimestamper

import (
	"flag"

	"github.com/stratumn/sdk/blockchain/btc"

	log "github.com/sirupsen/logrus"
)

var (
	fee int64
)

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.Int64Var(&fee, "fee", DefaultFee, "transaction fee (satoshis)")

}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to initialize
// a bcbatchfossilizer using flag values.
func InitializeWithFlags(version, commit string, key string, unspentFinder btc.UnspentFinder, broadcaster btc.Broadcaster) *Timestamper {
	ts, err := New(&Config{
		UnspentFinder: unspentFinder,
		Broadcaster:   broadcaster,
		WIF:           key,
		Fee:           fee,
	})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create Bitcoin timestamper")
	}
	return ts
}
