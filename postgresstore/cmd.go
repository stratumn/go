// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package postgresstore

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/utils"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
	noTableCode     = pq.ErrorCode("42P01")
)

var (
	create         bool
	drop           bool
	uniqueMapEntry bool
	url            string
)

// Initialize a postgres store adapter.
func Initialize(config *Config, create, drop, uniqueMapEntry bool) *Store {
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

	if uniqueMapEntry {
		err = store.AdapterConfig(a).EnforceUniqueMapEntry()
		if err != nil {
			log.WithField("uniqueMapEntry", err.Error()).Fatal("Unable to configure unique map entry")
		}
	}

	return a
}

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.BoolVar(&create, "create", false, "create tables and indexes then exit")
	flag.BoolVar(&drop, "drop", false, "drop tables and indexes then exit")
	flag.BoolVar(&uniqueMapEntry, "uniquemapentry", false, "enforce unicity of the first link in each process map")
	flag.StringVar(&url, "url", utils.OrStrings(os.Getenv("POSTGRESSTORE_URL"), DefaultURL), "URL of the PostgreSQL database")
}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to initialize
// a postgres adapter using flag values.
func InitializeWithFlags(version, commit string) *Store {
	config := &Config{URL: url, Version: version, Commit: commit}
	return Initialize(config, create, drop, uniqueMapEntry)
}
