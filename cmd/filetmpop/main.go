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

// The command filetmpop starts a tmpop node with a filestore.
package main

import (
	"flag"

	"github.com/stratumn/go-core/filestore"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/tendermint"
	"github.com/stratumn/go-core/tmpop"
	"github.com/stratumn/go-core/validation"
)

var (
	path    = flag.String("path", filestore.DefaultPath, "Path to directory where files are stored")
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	tendermint.RegisterFlags()
	monitoring.RegisterFlags()
	validation.RegisterFlags()

	monitoring.SetVersion(version, commit)
}

func main() {
	flag.Parse()

	a, err := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})
	if err != nil {
		monitoring.LogEntry().Fatal(err)
	}

	tmpopConfig := &tmpop.Config{
		Commit:     commit,
		Version:    version,
		Validation: validation.ConfigurationFromFlags(),
		Monitoring: monitoring.ConfigurationFromFlags(),
	}
	tmpop.Run(
		monitoring.WrapStore(a, "filestore"),
		monitoring.WrapKeyValueStore(a, "filestore"),
		tmpopConfig,
	)
}
