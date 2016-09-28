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

	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store/storehttp"

	"github.com/stratumn/goprivate/rethinkstore"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
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
	url      = flag.String("url", orStrings(os.Getenv("RETHINKSTORE_URL"), rethinkstore.DefaultURL), "URL of the RethinkDB database")
	db       = flag.String("db", orStrings(os.Getenv("RETHINKSTORE_DB"), rethinkstore.DefaultDB), "name of the RethinkDB database")
	hard     = flag.Bool("hard", rethinkstore.DefaultHard, "whether to use hard durability")
	certFile = flag.String("tlscert", "", "TLS certificate file")
	keyFile  = flag.String("tlskey", "", "TLS private key file")
	version  = "0.1.0"
	commit   = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()

	log.Infof("%s v%s@%s", rethinkstore.Description, version, commit[:7])
	log.Info("Copyright (c) 2016 Stratumn SAS")
	log.Info("All Rights Reserved")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	a, err := rethinkstore.New(&rethinkstore.Config{
		URL:     *url,
		DB:      *db,
		Hard:    *hard,
		Version: version,
		Commit:  commit,
	})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create RethinkDB store")
	}

	if *create {
		if err := a.Create(); err != nil {
			log.Fatalf("Fatal: %s", err)
		}
		log.WithField("error", err).Fatal("Failed to create RethinkDB tables and indexes")
		os.Exit(0)
	}

	if *drop {
		if err := a.Drop(); err != nil {
			log.WithField("error", err).Fatal("Failed to drop RethinkDB tables and indexes")
		}
		log.Info("Dropped tables and indexes")
		os.Exit(0)
	}

	exists := false
	for i := 1; i <= connectAttempts; i++ {
		if err != nil {
			time.Sleep(connectTimeout)
		}
		if exists, err = a.Exists(); err != nil {
			log.WithFields(log.Fields{
				"attempt": i,
				"max":     connectAttempts,
			}).Warn(fmt.Sprintf("Unable to connect to RethinkDB, retrying in %v", connectTimeout))
		} else {
			if !exists {
				if err = a.Create(); err != nil {
					log.WithField("error", err).Fatal("Failed to create RethinkDB tables and indexes")
				}
				log.Info("Created tables and indexes")
			}
			break
		}
	}
	if err != nil {
		log.WithField("max", connectAttempts).Fatal("Unable to connect to RethinkDB")
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
