// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/lib/pq"

	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store/storehttp"
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
	create   = flag.Bool("create", false, "create tables and indexes then exit")
	drop     = flag.Bool("drop", false, "drop tables and indexes then exit")
	http     = flag.String("http", storehttp.DefaultAddress, "HTTP address")
	url      = flag.String("url", orStrings(os.Getenv("POSTGRESSTORE_URL"), postgresstore.DefaultURL), "URL of the PostgreSQL database")
	certFile = flag.String("tlscert", "", "TLS certificate file")
	keyFile  = flag.String("tlskey", "", "TLS private key file")
	version  = "0.1.0"
	commit   = "00000000000000000000000000000000"
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

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.WithField("signal", sig).Info("Got exit signal")
		log.Info("Stopped")
		os.Exit(0)
	}()

	c := &jsonhttp.Config{
		Address:  *http,
		CertFile: *certFile,
		KeyFile:  *keyFile,
	}
	h := storehttp.New(a, c)

	log.WithField("http", *http).Info("Listening")
	if err := h.ListenAndServe(); err != nil {
		log.WithField("error", err).Fatal("Server stopped")
	}
}
