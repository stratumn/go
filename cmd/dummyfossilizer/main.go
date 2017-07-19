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

// The command dummnyfossilizer starts a fossilizerhttp server with a
// dummyfossilizer.
package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/sdk/dummyfossilizer"
	"github.com/stratumn/sdk/fossilizer/fossilizerhttp"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	fossilizerhttp.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", dummyfossilizer.Description, version, commit[:7])
	a := dummyfossilizer.New(&dummyfossilizer.Config{Version: version, Commit: commit})
	fossilizerhttp.RunWithFlags(a)
}
