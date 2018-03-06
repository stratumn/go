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

package elasticsearchstore

import (
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/utils"
)

var (
	sniffing bool
	url      string
)

// Initialize initializes a elasticsearch store adapter.
func Initialize(config *Config) *ESStore {
	var a *ESStore
	var storeErr error

	err := utils.Retry(func(attempt int) (retry bool, err error) {
		a, storeErr = New(config)

		if storeErr == nil {
			return false, nil
		}

		log.Infof("Unable to connect to elasticsearch. Retrying in 5s.")
		time.Sleep(5 * time.Second)
		return true, storeErr
	}, 10)

	if err != nil {
		log.Fatal(storeErr)
	}

	return a
}

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.StringVar(&url, "url", utils.OrStrings(os.Getenv("ELASTICSEARCH_URL"), DefaultURL), "URL of the ElasticSearch database")
	flag.BoolVar(&sniffing, "sniffing", false, "turn on elastic search nodes sniffing")
}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to initialize
// an elasticsearch adapter using flag values.
func InitializeWithFlags(version, commit string) *ESStore {
	config := &Config{URL: url, Version: version, Commit: commit, Sniffing: sniffing}
	return Initialize(config)

}
