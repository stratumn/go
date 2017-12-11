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

package rethinkstore

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/sdk/utils"
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

	exists, err := a.Exists()
	if err != nil {
		log.WithField("error", err).Fatal("Failed to check RethinkDB tables and indexes")
	}
	if !exists {
		if err = a.Create(); err != nil {
			log.WithField("error", err).Fatal("Failed to create RethinkDB tables and indexes")
		}
		log.Info("Created tables and indexes")
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
