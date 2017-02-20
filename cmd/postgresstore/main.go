// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"

	"github.com/stratumn/sdk/jsonhttp"
	"github.com/stratumn/sdk/jsonws"
	"github.com/stratumn/sdk/store/storehttp"
	"github.com/stratumn/goprivate/postgresstore"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
	noTableCode     = pq.ErrorCode("42P01")
)

func orStrings(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}

var (
	create          = flag.Bool("create", false, "create tables and indexes then exit")
	drop            = flag.Bool("drop", false, "drop tables and indexes then exit")
	url             = flag.String("url", orStrings(os.Getenv("POSTGRESSTORE_URL"), postgresstore.DefaultURL), "URL of the PostgreSQL database")
	http            = flag.String("http", storehttp.DefaultAddress, "HTTP address")
	wsReadBufSize   = flag.Int("wsreadbufsize", storehttp.DefaultWebSocketReadBufferSize, "Web socket read buffer size")
	wsWriteBufSize  = flag.Int("wswritebufsize", storehttp.DefaultWebSocketWriteBufferSize, "Web socket write buffer size")
	wsWriteChanSize = flag.Int("wswritechansize", storehttp.DefaultWebSocketWriteChanSize, "Size of a web socket connection write channel")
	wsWriteTimeout  = flag.Duration("wswritetimeout", storehttp.DefaultWebSocketWriteTimeout, "Timeout for a web socket write")
	wsPongTimeout   = flag.Duration("wspongtimeout", storehttp.DefaultWebSocketPongTimeout, "Timeout for a web socket expected pong")
	wsPingInterval  = flag.Duration("wspinginterval", storehttp.DefaultWebSocketPingInterval, "Interval between web socket pings")
	wsMaxMsgSize    = flag.Int64("maxmsgsize", storehttp.DefaultWebSocketMaxMsgSize, "Maximum size of a received web socket message")
	certFile        = flag.String("tlscert", "", "TLS certificate file")
	keyFile         = flag.String("tlskey", "", "TLS private key file")
	version         = "0.1.0"
	commit          = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()

	log.Infof("%s v%s@%s", postgresstore.Description, version, commit[:7])
	log.Info("Copyright (c) 2016 Stratumn SAS")
	log.Info("All Rights Reserved")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	a, err := postgresstore.New(&postgresstore.Config{URL: *url, Version: version, Commit: commit})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create PostgreSQL store")
	}

	if *create {
		if err := a.Create(); err != nil {
			log.WithField("error", err).Fatal("Failed to create PostgreSQL tables and indexes")
		}
		log.Info("Created tables and indexes")
		os.Exit(0)
	}

	if *drop {
		if err := a.Drop(); err != nil {
			log.WithField("error", err).Fatal("Failed to drop PostgreSQL tables and indexes")
		}
		log.Info("Dropped tables and indexes")
		os.Exit(0)
	}

	for i := 1; i <= connectAttempts; i++ {
		if err != nil {
			time.Sleep(connectTimeout)
		}
		if err = a.Prepare(); err != nil {
			if e, ok := err.(*pq.Error); ok && e.Code == noTableCode {
				if err = a.Create(); err != nil {
					log.WithField("error", err).Fatal("Failed to create PostgreSQL tables and indexes")
				}
				log.Info("Created tables and indexes")
			} else {
				log.WithFields(log.Fields{
					"attempt": i,
					"max":     connectAttempts,
				}).Warn(fmt.Sprintf("Unable to connect to PostgreSQL, retrying in %v", connectTimeout))
			}
		} else {
			break
		}
	}
	if err != nil {
		log.WithField("max", connectAttempts).Fatal("Unable to connect to PostgreSQL")
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
