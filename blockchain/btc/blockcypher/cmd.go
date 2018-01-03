package blockcypher

import (
	"context"
	"flag"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/stratumn/sdk/blockchain/btc"

	log "github.com/sirupsen/logrus"
)

var (
	bcyAPIKey       string
	limiterInterval time.Duration
	limiterSize     int
)

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.StringVar(&bcyAPIKey, "bcyapikey", "", "BlockCypher API key")
	flag.DurationVar(&limiterInterval, "limiterinterval", DefaultLimiterInterval, "BlockCypher API limiter interval")
	flag.IntVar(&limiterSize, "limitersize", DefaultLimiterSize, "BlockCypher API limiter size")
}

// RunWithFlags should be called after RegisterFlags and flag.Parse to initialize
// a blockcypher client using flag values.
func RunWithFlags(ctx context.Context, key string) *Client {
	if key == "" {
		log.Fatal("A WIF encoded private key is required")
	}

	WIF, err := btcutil.DecodeWIF(key)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to decode WIF encoded private key")
	}

	var network btc.Network
	if WIF.IsForNet(&chaincfg.TestNet3Params) {
		network = btc.NetworkTest3
	} else if WIF.IsForNet(&chaincfg.MainNetParams) {
		network = btc.NetworkMain
	} else {
		log.Fatal("WIF encoded private key uses unknown Bitcoin network")
	}

	bcy := New(&Config{
		Network:         network,
		APIKey:          bcyAPIKey,
		LimiterInterval: limiterInterval,
		LimiterSize:     limiterSize,
	})

	go bcy.Start(ctx)

	return bcy
}
