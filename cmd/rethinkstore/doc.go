// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// The command rethinkstore starts an HTTP server with a rethinkstore.
//
// Usage
//
// The following flags are available:
//
//	$ rethinkstore -h
//        -create
//          	create tables and indexes then exit
//        -db string
//          	name of the RethinkDB database (default "test")
//        -drop
//          	drop tables and indexes then exit
//        -hard
//          	whether to use hard durability (default true)
//        -http string
//          	HTTP address (default ":5000")
//        -tlscert string
//          	TLS certificate file
//        -tlskey string
//          	TLS private key file
//        -url string
//          	URL of the RethinkDB database (default "rethinkdb:28015")
//
// Env
//
//      RETHINKSTORE_URL="xxx" // URL of the RethinkDB database
//      RETHINKSTORE_DB="xxx"  // name of the RethinkDB database
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 5000:5000 stratumn/progresstore rethinkstore -url 'localhost:28015' -db test
package main
