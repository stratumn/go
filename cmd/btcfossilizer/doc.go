// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// The command btcfossilizer starts an HTTP server with a batch fossilizer on a Bitcoin blockchain.
//
// Usage
//
// The following flags are available:
//
//	$ btcfossilizer -h
//	Usage of btcfossilizer:
//	  -bcyapikey string
//	    	BlockCypher API key
//	  -callbacktimeout duration
//	    	callback requests timeout (default 10s)
//	  -fee int
//	    	transaction fee (satoshis) (default 15000)
//	  -interval duration
//	    	batch interval (default 1m0s)
//	  -maxleaves int
//	    	maximum number of leaves in a Merkle tree (default 32768)
//	  -port string
//	    	server port (default ":6000")
//	  -tlscert string
//	    	TLS certificate file
//	  -tlskey string
//	    	TLS private key file
//	  -verbose
//	    	verbose output
//	  -wif string
//	    	wallet import format key
//	  -workers int
//	    	number of result workers (default 8)
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 6000:6000 stratumn/btcfossilizer btcfossilizer -wif "your WIF key"
package main
