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

// The command rethinktmpop starts a tmpop node with a rethinkstore.
package main

import (
	"flag"

	"github.com/stratumn/sdk/rethinkstore"
	"github.com/stratumn/sdk/tendermint"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/validator"
)

var (
	cacheSize         = flag.Int("cacheSize", tmpop.DefaultCacheSize, "size of the cache of the storage tree")
	validatorFilename = flag.String("rules_filename", validator.DefaultFilename, "Path to filename containing validation rules")
	version           = "0.1.0"
	commit            = "00000000000000000000000000000000"
)

func init() {
	tendermint.RegisterFlags()
	rethinkstore.RegisterFlags()
}

func main() {
	flag.Parse()

	a := rethinkstore.InitializeWithFlags(version, commit)

	tmpopConfig := &tmpop.Config{Commit: commit, Version: version, CacheSize: *cacheSize, ValidatorFilename: *validatorFilename}

	tmpop.Run(a, tmpopConfig)
}
