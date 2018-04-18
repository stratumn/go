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

// The command filestore starts a storehttp server with a couchstore.

package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/couchstore"
	_ "github.com/stratumn/go-indigocore/fossilizer/evidences"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/store/storehttp"
	"github.com/stratumn/go-indigocore/utils"
)

var (
	endpoint = flag.String("endpoint", "http://localhost:5984", "CouchDB endpoint")
	version  = "x.x.x"
	commit   = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	monitoring.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", couchstore.Description, version, commit[:7])

	var a *couchstore.CouchStore
	var storeErr error

	err := utils.Retry(func(attempt int) (retry bool, err error) {
		a, storeErr = couchstore.New(&couchstore.Config{
			Address: *endpoint,
			Version: version,
			Commit:  commit,
		})

		if storeErr == nil {
			return false, nil
		}

		if _, ok := storeErr.(*couchstore.CouchNotReadyError); ok {
			log.Infof("Unable to connect to couchdb (%v). Retrying in 5s.", storeErr.Error())
			time.Sleep(5 * time.Second)
			return true, storeErr
		}

		return false, storeErr
	}, 10)

	if err != nil {
		log.Fatal(storeErr)
	}

	storehttp.RunWithFlags(monitoring.NewStoreAdapter(a, "couchstore"))
}
