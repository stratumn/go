// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command dummystore starts a storehttp server with a dummystore.
package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/dummystore"
	"github.com/stratumn/sdk/store/storehttp"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", dummystore.Description, version, commit[:7])
	a := dummystore.New(&dummystore.Config{Version: version, Commit: commit})
	storehttp.RunWithFlags(a)
}
