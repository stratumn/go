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
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/stratumn/sdk/tendermint"
	"github.com/tendermint/abci/types"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	tendermint.RegisterFlags()
}

func main() {
	flag.Parse()

	log.Infof("TMDummy v%s@%s", version, commit[:7])
	log.Info("Copyright (c) 2017 Stratumn SAS")
	log.Info("Apache License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	tendermint.RunNodeForever(tendermint.GetConfig(), types.NewBaseApplication())
}
