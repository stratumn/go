// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// The command dummybatchfossilizer starts an HTTP server with a batch fossilizer on a dummy blockchain.
//
// Usage
//
// The following flags are available:
//
//	$ dummybatchfossilizer -h
//	Usage of dummybatchfossilizer:
//        -archive
//          	whether to archive completed batches (requires path) (default true)
//        -callbacktimeout duration
//          	callback requests timeout (default 10s)
//        -exitbatch
//          	whether to do a batch on exit (default true)
//        -fsync
//          	whether to fsync after saving a pending hash
//        -http string
//          	HTTP address (default ":6000")
//        -interval duration
//          	batch interval (default 10m0s)
//        -maxleaves int
//          	maximum number of leaves in a Merkle tree (default 32768)
//        -path string
//          	an optional path to store files
//        -tlscert string
//          	TLS certificate file
//        -tlskey string
//          	TLS private key file
//        -workers int
//          	number of result workers (default 8)
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 6000:6000 stratumn/dummybatchfossilizer
package main
