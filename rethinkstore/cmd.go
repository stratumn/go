// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package rethinkstore

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stratumn/goprivate/utils"

	log "github.com/sirupsen/logrus"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
)

var (
	create bool
	drop   bool
	url    string
	db     string
	hard   bool
)

// Initialize initializes a rethinkdb store adapter
func Initialize(config *Config, create, drop bool) *Store {
	a, err := New(config)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create RethinkDB store")
	}

	if create {
		if err := a.Create(); err != nil {
			log.Fatalf("Fatal: %s", err)
		}
		log.WithField("error", err).Fatal("Failed to create RethinkDB tables and indexes")
		os.Exit(0)
	}

	if drop {
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
	return a
}

// RegisterFlags register the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.BoolVar(&create, "create", false, "create tables and indexes then exit")
	flag.BoolVar(&drop, "drop", false, "drop tables and indexes then exit")
	flag.StringVar(&url, "url", utils.OrStrings(os.Getenv("RETHINKSTORE_URL"), DefaultURL), "URL of the RethinkDB database")
	flag.StringVar(&db, "db", utils.OrStrings(os.Getenv("RETHINKSTORE_DB"), DefaultDB), "name of the RethinkDB database")
	flag.BoolVar(&hard, "hard", DefaultHard, "whether to use hard durability")
}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to intialize
// a rethinkdb adapter using flag values.
func InitializeWithFlags(version, commit string) *Store {
	config := &Config{
		URL:     url,
		DB:      db,
		Hard:    hard,
		Version: version,
		Commit:  commit,
	}
	return Initialize(config, create, drop)
}