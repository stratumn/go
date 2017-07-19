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
