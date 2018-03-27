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

// The command filetmpop starts a tmpop node with a filestore.
package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	_ "github.com/stratumn/go-indigocore/cs/evidences"
	"github.com/stratumn/go-indigocore/filestore"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/tendermint"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stratumn/go-indigocore/validator"
)

var (
	path              = flag.String("path", filestore.DefaultPath, "Path to directory where files are stored")
	validatorFilename = flag.String("rules_filename", validator.DefaultFilename, "Path to filename containing validation rules")
	version           = "x.x.x"
	commit            = "00000000000000000000000000000000"
)

func init() {
	tendermint.RegisterFlags()
	monitoring.RegisterFlags()
}

func main() {
	flag.Parse()

	a, err := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})
	if err != nil {
		log.Fatal(err)
	}

	tmpopConfig := &tmpop.Config{
		Commit:            commit,
		Version:           version,
		ValidatorFilename: *validatorFilename,
		Monitoring:        monitoring.ConfigurationFromFlags(),
	}
	tmpop.Run(a, a, tmpopConfig)
}
