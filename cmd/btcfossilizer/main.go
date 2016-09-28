// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"

	"github.com/stratumn/go/fossilizer/fossilizerhttp"
	"github.com/stratumn/go/jsonhttp"

	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/bcbatchfossilizer"
	"github.com/stratumn/goprivate/blockchain/btc"
	"github.com/stratumn/goprivate/blockchain/btc/blockcypher"
	"github.com/stratumn/goprivate/blockchain/btc/btctimestamper"
	"github.com/stratumn/goprivate/merkle"
)

var (
	http             = flag.String("http", fossilizerhttp.DefaultAddress, "HTTP address")
	certFile         = flag.String("tlscert", "", "TLS certificate file")
	keyFile          = flag.String("tlskey", "", "TLS private key file")
	numResultWorkers = flag.Int("workers", fossilizerhttp.DefaultNumResultWorkers, "number of result workers")
	callbackTimeout  = flag.Duration("callbacktimeout", fossilizerhttp.DefaultCallbackTimeout, "callback requests timeout")
	interval         = flag.Duration("interval", batchfossilizer.DefaultInterval, "batch interval")
	maxLeaves        = flag.Int("maxleaves", batchfossilizer.DefaultMaxLeaves, "maximum number of leaves in a Merkle tree")
	path             = flag.String("path", "", "an optional path to store files")
	archive          = flag.Bool("archive", batchfossilizer.DefaultArchive, "whether to archive completed batches (requires path)")
	exitBatch        = flag.Bool("exitbatch", batchfossilizer.DefaultStopBatch, "whether to do a batch on exit")
	fsync            = flag.Bool("fsync", batchfossilizer.DefaultFSync, "whether to fsync after saving a pending hash (requires path)")
	key              = flag.String("wif", os.Getenv("BTCFOSSILIZER_WIF"), "wallet import format key")
	fee              = flag.Int64("fee", btctimestamper.DefaultFee, "transaction fee (satoshis)")
	bcyAPIKey        = flag.String("bcyapikey", "", "BlockCypher API key")
	limiterInterval  = flag.Duration("limiterinterval", blockcypher.DefaultLimiterInterval, "BlockCypher API limiter interval")
	limiterSize      = flag.Int("limitersize", blockcypher.DefaultLimiterSize, "BlockCypher API limiter size")
	version          = "0.1.0"
	commit           = "00000000000000000000000000000000"
)

func main() {

	flag.Parse()

	if *key == "" {
		log.Fatal("A WIF encoded private key is required")
	}

	WIF, err := btcutil.DecodeWIF(*key)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to decode WIF encoded private key")
	}

	var network btc.Network
	if WIF.IsForNet(&chaincfg.TestNet3Params) {
		network = btc.NetworkTest3
	} else if WIF.IsForNet(&chaincfg.MainNetParams) {
		network = btc.NetworkMain
	} else {
		log.Fatal("WIF encoded private key uses nknown Bitcoin network")
	}

	log.Infof("%s v%s@%s", bcbatchfossilizer.Description, version, commit[:7])
	log.Info("Copyright (c) 2016 Stratumn SAS")
	log.Info("All Rights Reserved")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	bcy := blockcypher.New(&blockcypher.Config{
		Network:         network,
		APIKey:          *bcyAPIKey,
		LimiterInterval: *limiterInterval,
		LimiterSize:     *limiterSize,
	})
	ts, err := btctimestamper.New(&btctimestamper.Config{
		UnspentFinder: bcy,
		Broadcaster:   bcy,
		WIF:           *key,
		Fee:           *fee,
	})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create Bitcoin timestamper")
	}

	a, err := bcbatchfossilizer.New(&bcbatchfossilizer.Config{
		HashTimestamper: ts,
	}, &batchfossilizer.Config{
		Version:   version,
		Commit:    commit,
		Interval:  *interval,
		MaxLeaves: *maxLeaves,
		Path:      *path,
		Archive:   *archive,
		StopBatch: *exitBatch,
		FSync:     *fsync,
	})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create blockchain batch fossilizer")
	}

	go func() {
		if err := a.Start(); err != nil {
			log.WithField("error", err).Fatal("Failed to start blockchain batch fossilizer")
		}
	}()

	go bcy.Start()

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.WithField("signal", sig).Info("Got exit signal")
		log.Info("Cleaning up")
		a.Stop()
		bcy.Stop()
		log.Info("Stopped")
		os.Exit(0)
	}()

	config := &fossilizerhttp.Config{
		NumResultWorkers: *numResultWorkers,
		CallbackTimeout:  *callbackTimeout,
		MinDataLen:       merkle.HashByteSize * 2,
		MaxDataLen:       merkle.HashByteSize * 2,
	}
	httpConfig := &jsonhttp.Config{
		Address:  *http,
		CertFile: *certFile,
		KeyFile:  *keyFile,
	}
	h := fossilizerhttp.New(a, config, httpConfig)

	log.WithField("http", *http).Info("Listening")
	if err := h.ListenAndServe(); err != nil {
		log.WithField("error", err).Fatal("Server stopped")
	}
}
