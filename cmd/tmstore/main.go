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

package main

import (
	"flag"

	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/jsonws"
	"github.com/stratumn/go/store/storehttp"
	"github.com/stratumn/go/tmstore"
	"github.com/stratumn/go/utils"
)

var (
	http              = flag.String("http", storehttp.DefaultAddress, "HTTP address")
	wsReadBufSize     = flag.Int("wsreadbufsize", storehttp.DefaultWebSocketReadBufferSize, "Web socket read buffer size")
	wsWriteBufSize    = flag.Int("wswritebufsize", storehttp.DefaultWebSocketWriteBufferSize, "Web socket write buffer size")
	wsWriteChanSize   = flag.Int("wswritechansize", storehttp.DefaultWebSocketWriteChanSize, "Size of a web socket connection write channel")
	wsWriteTimeout    = flag.Duration("wswritetimeout", storehttp.DefaultWebSocketWriteTimeout, "Timeout for a web socket write")
	wsPongTimeout     = flag.Duration("wspongtimeout", storehttp.DefaultWebSocketPongTimeout, "Timeout for a web socket expected pong")
	wsPingInterval    = flag.Duration("wspinginterval", storehttp.DefaultWebSocketPingInterval, "Interval between web socket pings")
	wsMaxMsgSize      = flag.Int64("maxmsgsize", storehttp.DefaultWebSocketMaxMsgSize, "Maximum size of a received web socket message")
	endpoint          = flag.String("endpoint", tmstore.DefaultEndpoint, "endpoint used to communicate with Tendermint Core")
	tmWsRetryInterval = flag.Duration("tmWsRetryInterval", tmstore.DefaultWsRetryInterval, "interval between tendermint websocket connection tries")
	certFile          = flag.String("tlscert", "", "TLS certificate file")
	keyFile           = flag.String("tlskey", "", "TLS private key file")
	version           = "0.1.0"
	commit            = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", tmstore.Description, version, commit[:7])

	a := tmstore.New(&tmstore.Config{Endpoint: *endpoint, Version: version, Commit: commit})

	go utils.Retry(func(attempt int) (retry bool, err error) {
		err = a.StartWebsocket()
		if err != nil {
			log.Infof("%v, retrying...", err)
			time.Sleep(*tmWsRetryInterval)
		}
		return true, err
	}, 0)

	httpConfig := &jsonhttp.Config{
		Address:  *http,
		CertFile: *certFile,
		KeyFile:  *keyFile,
	}
	basicConfig := &jsonws.BasicConfig{
		ReadBufferSize:  *wsReadBufSize,
		WriteBufferSize: *wsWriteBufSize,
	}
	bufConnConfig := &jsonws.BufferedConnConfig{
		Size:         *wsWriteChanSize,
		WriteTimeout: *wsWriteTimeout,
		PongTimeout:  *wsPongTimeout,
		PingInterval: *wsPingInterval,
		MaxMsgSize:   *wsMaxMsgSize,
	}
	storehttp.Run(a, httpConfig, basicConfig, bufConnConfig)
}
