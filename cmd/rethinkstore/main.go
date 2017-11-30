// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

// The command rethinkstore starts an HTTP server with a rethinkstore.
package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/sdk/store/storehttp"

	"github.com/stratumn/goprivate/rethinkstore"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	rethinkstore.RegisterFlags()
}

func main() {
	flag.Parse()

	log.Infof("%s v%s@%s", rethinkstore.Description, version, commit[:7])

	a := rethinkstore.InitializeWithFlags(version, commit)
	storehttp.RunWithFlags(a)
}
