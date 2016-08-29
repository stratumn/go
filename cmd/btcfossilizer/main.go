// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

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
	port             = flag.String("port", fossilizerhttp.DefaultPort, "server port")
	certFile         = flag.String("tlscert", "", "TLS certificate file")
	keyFile          = flag.String("tlskey", "", "TLS private key file")
	numResultWorkers = flag.Int("workers", fossilizerhttp.DefaultNumResultWorkers, "number of result workers")
	callbackTimeout  = flag.Duration("callbacktimeout", fossilizerhttp.DefaultCallbackTimeout, "callback requests timeout")
	verbose          = flag.Bool("verbose", fossilizerhttp.DefaultVerbose, "verbose output")
	interval         = flag.Duration("interval", batchfossilizer.DefaultInterval, "batch interval")
	maxLeaves        = flag.Int("maxleaves", batchfossilizer.DefaultMaxLeaves, "maximum number of leaves in a Merkle tree")
	path             = flag.String("path", "", "an optional path to store files")
	archive          = flag.Bool("archive", batchfossilizer.DefaultArchive, "whether to archive completed batches (requires path)")
	exitBatch        = flag.Bool("exitbatch", batchfossilizer.DefaultStopBatch, "whether to do a batch on exit")
	fsync            = flag.Bool("fsync", batchfossilizer.DefaultFSync, "whether to fsync after saving a pending hash (requires path)")
	key              = flag.String("wif", "", "wallet import format key")
	fee              = flag.Int64("fee", btctimestamper.DefaultFee, "transaction fee (satoshis)")
	bcyAPIKey        = flag.String("bcyapikey", "", "BlockCypher API key")
	version          = "0.1.0"
	commit           = "00000000000000000000000000000000"
)

func main() {

	flag.Parse()

	if *key == "" {
		log.Fatal("Fatal: a WIF encoded private key is required")
	}

	WIF, err := btcutil.DecodeWIF(*key)
	if err != nil {
		log.Fatalf("Fatal: %s", err)
	}

	var network btc.Network
	if WIF.IsForNet(&chaincfg.TestNet3Params) {
		network = btc.NetworkTest3
	} else if WIF.IsForNet(&chaincfg.MainNetParams) {
		network = btc.NetworkMain
	} else {
		log.Fatal("Fatal: unknown Bitcoin network")
	}

	log.SetPrefix(fmt.Sprintf("btcfossilizer:%s ", network))

	log.Printf("%s v%s@%s", bcbatchfossilizer.Description, version, commit[:6])
	log.Print("Copyright (c) 2016 Stratumn SAS")
	log.Print("All Rights Reserved")
	log.Printf("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	bcy := blockcypher.New(network, *bcyAPIKey)
	ts, err := btctimestamper.New(&btctimestamper.Config{
		UnspentFinder: bcy,
		Broadcaster:   bcy,
		WIF:           *key,
		Fee:           *fee,
	})
	if err != nil {
		log.Fatalf("Fatal: %s", err)
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
		log.Fatalf("Fatal: %s", err)
	}

	go func() {
		if err := a.Start(); err != nil {
			log.Fatalf("Fatal: %s", err)
		}
	}()

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.Printf("Got signal %q", sig)
		log.Print("Cleaning up")
		if err := a.Stop(); err != nil {
			log.Printf("Error: %s", err)
			os.Exit(1)
		}
		log.Print("Stopped")
		os.Exit(0)
	}()

	c := &fossilizerhttp.Config{
		Config: jsonhttp.Config{
			Port:     *port,
			CertFile: *certFile,
			KeyFile:  *keyFile,
			Verbose:  *verbose,
		},
		NumResultWorkers: *numResultWorkers,
		CallbackTimeout:  *callbackTimeout,
		MinDataLen:       merkle.HashByteSize * 2,
		MaxDataLen:       merkle.HashByteSize * 2,
	}
	h := fossilizerhttp.New(a, c)

	log.Printf("Listening on %q", *port)
	if err := h.ListenAndServe(); err != nil {
		log.Fatalf("Fatal: %s", err)
	}
}
