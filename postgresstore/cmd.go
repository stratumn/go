// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package postgresstore

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/lib/pq"

	"github.com/stratumn/sdk/utils"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
	noTableCode     = pq.ErrorCode("42P01")
)

var (
	create bool
	drop   bool
	url    string
)

// Initialize initializes a postgres store adapter
func Initialize(config *Config, create, drop bool) *Store {
	a, err := New(config)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create PostgreSQL store")
	}

	if create {
		if err := a.Create(); err != nil {
			log.WithField("error", err).Fatal("Failed to create PostgreSQL tables and indexes")
		}
		log.Info("Created tables and indexes")
		os.Exit(0)
	}

	if drop {
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
	return a
}

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.BoolVar(&create, "create", false, "create tables and indexes then exit")
	flag.BoolVar(&drop, "drop", false, "drop tables and indexes then exit")
	flag.StringVar(&url, "url", utils.OrStrings(os.Getenv("POSTGRESSTORE_URL"), DefaultURL), "URL of the PostgreSQL database")
}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to intialize
// a postgres adapter using flag values.
func InitializeWithFlags(version, commit string) *Store {
	config := &Config{URL: url, Version: version, Commit: commit}
	return Initialize(config, create, drop)

}
