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
	"os"
	"time"

	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/util"
)

const (
	connectAttempts = 15
	connectTimeout  = 5 * time.Second
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
		monitoring.LogEntry().WithField("error", err).Fatal("Failed to create PostgreSQL store")
	}

	if drop {
		dropDB(a)
		os.Exit(0)
	}

	// Ensure the DB tables are created.
	createDB(a)

	if create {
		os.Exit(0)
	}

	prepareDB(a)

	if uniqueMapEntry {
		err = store.AdapterConfig(a).EnforceUniqueMapEntry()
		if err != nil {
			monitoring.LogEntry().WithField("uniqueMapEntry", err.Error()).Fatal("Unable to configure unique map entry.")
		}
	}

	return a
}

// createDB creates schemas, tables and indexes.
func createDB(a *Store) {
	err := util.Retry(func(attempt int) (bool, error) {
		if err := a.Create(); err != nil {
			monitoring.LogEntry().WithField("error", err).Warn("Failed to create PostgreSQL tables and indexes. Retrying...")
			time.Sleep(connectTimeout)
			return true, err
		}

		return false, nil
	}, connectAttempts)

	if err != nil {
		monitoring.LogEntry().WithField("error", err).Fatal("Failed to create PostgreSQL tables and indexes.")
	}

	monitoring.LogEntry().Info("Created tables and indexes.")
}

// prepareDB prepares statements.
func prepareDB(a *Store) {
	err := util.Retry(func(attempt int) (bool, error) {
		if err := a.Prepare(); err != nil {
			monitoring.LogEntry().WithField("error", err).Warn("Failed to prepare PostgreSQL statements. Retrying...")
			time.Sleep(connectTimeout)
			return true, err
		}

		return false, nil
	}, connectAttempts)

	if err != nil {
		monitoring.LogEntry().WithField("error", err).Fatal("Failed to prepare PostgreSQL statements.")
	}

	monitoring.LogEntry().Info("Prepared PostgreSQL statements.")
}

// dropDB drops schemas, tables and indexes.
func dropDB(a *Store) {
	err := util.Retry(func(attempt int) (bool, error) {
		if err := a.Drop(); err != nil {
			monitoring.LogEntry().WithField("error", err).Warn("Failed to drop PostgreSQL tables and indexes. Retrying...")
			time.Sleep(connectTimeout)
			return true, err
		}

		return false, nil
	}, connectAttempts)

	if err != nil {
		monitoring.LogEntry().WithField("error", err).Fatal("Failed to drop PostgreSQL tables and indexes.")
	}

	monitoring.LogEntry().Info("Dropped tables and indexes.")
}

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.BoolVar(&create, "create", false, "create tables and indexes then exit")
	flag.BoolVar(&drop, "drop", false, "drop tables and indexes then exit")
	flag.BoolVar(&uniqueMapEntry, "uniquemapentry", false, "enforce unicity of the first link in each process map")
	flag.StringVar(&url, "url", DefaultURL, "URL of the PostgreSQL database (should be set via the POSTGRESSTORE_URL environment variable)")
}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to initialize
// a postgres adapter using flag values.
func InitializeWithFlags(version, commit string) *Store {
	dbURL := util.OrStrings(os.Getenv("POSTGRESSTORE_URL"), DefaultURL)

	config := &Config{URL: dbURL, Version: version, Commit: commit}
	return Initialize(config, create, drop, uniqueMapEntry)
}
