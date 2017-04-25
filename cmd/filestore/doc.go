// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command filestore starts an HTTP server with a filestore.
//
// Usage:
//
//	$ filestore -h
//	Usage of filestore:
//	  -didsavechansize int
//	    	Size of the DidSave channel (default 256)
//	  -http string
//	    	HTTP address (default ":5000")
//	  -maxheaderbytes int
//	    	maximum header bytes (default 256)
//	  -maxmsgsize int
//	    	Maximum size of a received web socket message (default 32768)
//	  -path string
//	    	path to directory where files are stored (default "/var/stratumn/filestore")
//	  -readtimeout duration
//	    	read timeout (default 10s)
//	  -shutdowntimeout duration
//	    	shutdown timeout (default 10s)
//	  -tlscert string
//	    	TLS certificate file
//	  -tlskey string
//	    	TLS private key file
//	  -writetimeout duration
//	    	write timeout (default 10s)
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
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 5000:5000 stratumn/filestore
package main
