// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/filestore"
	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/jsonws"
	"github.com/stratumn/sdk/store/storehttp"
)

var (
	http            = flag.String("http", storehttp.DefaultAddress, "HTTP address")
	wsReadBufSize   = flag.Int("wsreadbufsize", storehttp.DefaultWebSocketReadBufferSize, "Web socket read buffer size")
	wsWriteBufSize  = flag.Int("wswritebufsize", storehttp.DefaultWebSocketWriteBufferSize, "Web socket write buffer size")
	wsWriteChanSize = flag.Int("wswritechansize", storehttp.DefaultWebSocketWriteChanSize, "Size of a web socket connection write channel")
	wsWriteTimeout  = flag.Duration("wswritetimeout", storehttp.DefaultWebSocketWriteTimeout, "Timeout for a web socket write")
	wsPongTimeout   = flag.Duration("wspongtimeout", storehttp.DefaultWebSocketPongTimeout, "Timeout for a web socket expected pong")
	wsPingInterval  = flag.Duration("wspinginterval", storehttp.DefaultWebSocketPingInterval, "Interval between web socket pings")
	wsMaxMsgSize    = flag.Int64("maxmsgsize", storehttp.DefaultWebSocketMaxMsgSize, "Maximum size of a received web socket message")
	path            = flag.String("path", filestore.DefaultPath, "path to directory where files are stored")
	certFile        = flag.String("tlscert", "", "TLS certificate file")
	keyFile         = flag.String("tlskey", "", "TLS private key file")
	version         = "0.1.0"
	commit          = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", filestore.Description, version, commit[:7])

	a, err := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})
	if err != nil {
		log.Fatal(err)
	}

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
