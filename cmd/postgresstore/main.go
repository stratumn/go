// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

// The command postgresstore starts an HTTP server with a postgresstore.
package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/sdk/postgresstore"
	"github.com/stratumn/sdk/store/storehttp"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	postgresstore.RegisterFlags()
}

func main() {
	flag.Parse()

	log.Infof("%s v%s@%s", postgresstore.Description, version, commit[:7])

	a := postgresstore.InitializeWithFlags(version, commit)
	storehttp.RunWithFlags(a)
}
