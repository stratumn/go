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

// The command filestore starts a storehttp server with a filestore.
package main

import (
	"flag"

	"github.com/stratumn/go-core/filestore"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/store/storehttp"
	"github.com/stratumn/go-core/validation"
)

var (
	path    = flag.String("path", filestore.DefaultPath, "Path to directory where files are stored")
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
	monitoring.RegisterFlags()
	validation.RegisterFlags()

	monitoring.SetVersion(version, commit)
}

func main() {
	flag.Parse()

	monitoring.LogEntry().Infof("%s v%s@%s", filestore.Description, version, commit[:7])

	var err error
	var a store.Adapter

	a, err = filestore.New(&filestore.Config{
		Path:    *path,
		Version: version,
		Commit:  commit,
	})
	if err != nil {
		monitoring.LogEntry().Fatal(err)
	}

	a, err = validation.WrapStoreWithConfigFile(a, validation.ConfigurationFromFlags())
	if err != nil {
		monitoring.LogEntry().Fatal(err)
	}

	storehttp.RunWithFlags(monitoring.WrapStore(a, "filestore"))
}
