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

// The command postgresstore starts an HTTP server with a postgresstore.
package main

import (
	"flag"

	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/postgresstore"
	"github.com/stratumn/go-core/store/storehttp"
	"github.com/stratumn/go-core/validation"
)

var (
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	postgresstore.RegisterFlags()
	monitoring.RegisterFlags()
	validation.RegisterFlags()

	monitoring.SetVersion(version, commit)
}

func main() {
	flag.Parse()

	monitoring.LogEntry().Infof("%s v%s@%s", postgresstore.Description, version, commit[:7])

	a, err := validation.WrapStoreWithConfigFile(
		postgresstore.InitializeWithFlags(version, commit),
		validation.ConfigurationFromFlags(),
	)
	if err != nil {
		monitoring.LogEntry().Fatal(err)
	}

	storehttp.RunWithFlags(monitoring.WrapStore(a, "postgresstore"))
}
