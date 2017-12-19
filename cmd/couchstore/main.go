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

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/sdk/couchstore"
	_ "github.com/stratumn/sdk/cs/evidences"
	"github.com/stratumn/sdk/store/storehttp"
)

var (
	endpoint = flag.String("endpoint", "http://localhost:5984", "CouchDB endpoint")
	version  = "x.x.x"
	commit   = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", couchstore.Description, version, commit[:7])

	a, err := couchstore.New(&couchstore.Config{
		Address: *endpoint,
		Version: version,
		Commit:  commit,
	})
	if err != nil {
		log.Fatal(err)
	}
	storehttp.RunWithFlags(a)
}
