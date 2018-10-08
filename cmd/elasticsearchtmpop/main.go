// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

// The command elasticsearchtmpop starts a tmpop node with a elasticsearchstore.
package main

import (
	"flag"

	"github.com/stratumn/go-core/elasticsearchstore"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/tendermint"
	"github.com/stratumn/go-core/tmpop"
	"github.com/stratumn/go-core/validation"
)

var (
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	tendermint.RegisterFlags()
	elasticsearchstore.RegisterFlags()
	monitoring.RegisterFlags()
	validation.RegisterFlags()
}

func main() {
	flag.Parse()

	a := elasticsearchstore.InitializeWithFlags(version, commit)
	tmpopConfig := &tmpop.Config{
		Commit:     commit,
		Version:    version,
		Validation: validation.ConfigurationFromFlags(),
		Monitoring: monitoring.ConfigurationFromFlags(),
	}
	tmpop.Run(
		monitoring.WrapStore(a, "elasticsearchstore"),
		monitoring.WrapKeyValueStore(a, "elasticsearchstore"),
		tmpopConfig,
	)
}
