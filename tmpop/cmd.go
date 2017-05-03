// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"runtime"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tendermint"
)

// Run launches a TMPop Tendermint App
func Run(a store.Adapter, config *Config) {
	adapterInfo, err := a.GetInfo()
	if err != nil {
		log.Fatal(err)
	}

	tmpop, err := New(a, config)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("TMPop v%s@%s", config.Version, config.Commit[:7])
	log.Infof("Adapter %v", adapterInfo)
	log.Info("Copyright (c) 2017 Stratumn SAS")
	log.Info("Mozilla Public License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	tendermint.RunNodeForever(tendermint.GetConfig(), tmpop)
}
