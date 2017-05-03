// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command tmstore starts an HTTP server with a tmstore.
package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/store/storehttp"
	"github.com/stratumn/sdk/tmstore"
)

var (
	endpoint          = flag.String("endpoint", tmstore.DefaultEndpoint, "Endpoint used to communicate with Tendermint Core")
	tmWsRetryInterval = flag.Duration("tm_ws_retry_interval", tmstore.DefaultWsRetryInterval, "Interval between tendermint websocket connection tries")
	version           = "0.1.0"
	commit            = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", tmstore.Description, version, commit[:7])

	a := tmstore.New(&tmstore.Config{Endpoint: *endpoint, Version: version, Commit: commit})
	go a.RetryStartWebsocket(*tmWsRetryInterval)

	storehttp.RunWithFlags(a)
}
