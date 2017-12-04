// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The command btcfossilizer starts an HTTP server with a batch fossilizer on a
// Bitcoin blockchain.
//
// Usage
//
// The following flags are available:
//
//	$ btcfossilizer -h
//	Usage of btcfossilizer:
//        -archive
//          	whether to archive completed batches (requires path) (default true)
//        -bcyapikey string
//          	BlockCypher API key
//        -callbacktimeout duration
//          	callback requests timeout (default 10s)
//        -exitbatch
//          	whether to do a batch on exit (default true)
//        -fee int
//          	transaction fee (satoshis) (default 15000)
//        -fsync
//          	whether to fsync after saving a pending hash (requires path)
//        -http string
//          	HTTP address (default ":6000")
//        -interval duration
//          	batch interval (default 10m0s)
//        -limiterinterval duration
//          	BlockCypher API limiter interval (default 1m0s)
//        -limitersize int
//          	BlockCypher API limiter size (default 2)
//        -maxleaves int
//          	maximum number of leaves in a Merkle tree (default 32768)
//        -path string
//          	an optional path to store files
//        -tlscert string
//          	TLS certificate file
//        -tlskey string
//          	TLS private key file
//        -wif string
//          	wallet import format key
//        -workers int
//          	number of result workers (default 8)
//
// Env
//
//      BTCFOSSILIZER_WIF="xxx" // wallet import format key
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 6000:6000 stratumn/btcfossilizer btcfossilizer \
//              -wif "your WIF key"
package main
