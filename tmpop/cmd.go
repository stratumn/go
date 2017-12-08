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

package tmpop

import (
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tendermint"
)

// Run launches a TMPop Tendermint App
func Run(a store.Adapter, kv store.KeyValueStore, config *Config) {
	adapterInfo, err := a.GetInfo()
	if err != nil {
		log.Fatal(err)
	}

	tmpop, err := New(a, kv, config)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("TMPop v%s@%s", config.Version, config.Commit[:7])
	log.Infof("Adapter %v", adapterInfo)
	log.Info("Copyright (c) 2017 Stratumn SAS")
	log.Info("Apache License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	tendermintNode := tendermint.NewNode(tendermint.GetConfig(), tmpop)
	tendermintClient := NewTendermintClient(client.NewLocal(tendermintNode))
	tmpop.ConnectTendermint(tendermintClient)
	tendermintNode.Start()
	tendermintNode.RunForever()
}
