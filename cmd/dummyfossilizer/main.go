// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command dummnyfossilizer starts a fossilizerhttp server with a
// dummyfossilizer.
package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/sdk/dummyfossilizer"
	"github.com/stratumn/sdk/fossilizer/fossilizerhttp"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	fossilizerhttp.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", dummyfossilizer.Description, version, commit[:7])
	a := dummyfossilizer.New(&dummyfossilizer.Config{Version: version, Commit: commit})
	fossilizerhttp.RunWithFlags(a)
}
