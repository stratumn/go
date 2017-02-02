// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

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
//	  -maxmsgsize int
//	    	Maximum size of a received web socket message (default 32768)
//        -tlscert string
//          	TLS certificate file
//        -tlskey string
//          	TLS private key file
//        -url string
//          	URL of the RethinkDB database (default "rethinkdb:28015")
//	  -wspinginterval duration
//	    	Interval between web socket pings (default 54s)
//	  -wspongtimeout duration
//	    	Timeout for a web socket expected pong (default 1m0s)
//	  -wsreadbufsize int
//	    	Web socket read buffer size (default 1024)
//	  -wswritebufsize int
//	    	Web socket write buffer size (default 1024)
//	  -wswritechansize int
//	    	Size of a web socket connection write channel (default 256)
//	  -wswritetimeout duration
//	    	Timeout for a web socket write (default 10s)
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
//	$ docker run -p 5000:5000 stratumn/progresstore rethinkstore \
//              -url 'localhost:28015' -db test
package main
