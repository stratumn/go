// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// The command postgresstore starts an HTTP server with a postgresstore.
//
// Usage
//
// The following flags are available:
//
//	$ postgresstore -h
//    -create
//      	create tables and indexes then exit
//    -drop
//      	drop tables and indexes then exit
//    -port string
//      	server port (default ":5000")
//    -tlscert string
//      	TLS certificate file
//    -tlskey string
//      	TLS private key file
//    -url string
//      	URL of the PostgreSQL database (default "postgres://postgres@postgres/postgres?sslmode=disable")
//    -verbose
//      	verbose output
//
// Env
//
//      POSTGRESSTORE_URL="xxx" // URL of the PostgreSQL database
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 5000:5000 stratumn/progresstore postgresstore -url 'postgres://postgres@localhost/postgres?sslmode=disable'
package main
