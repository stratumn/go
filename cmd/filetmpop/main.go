// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/filestore"
	"github.com/stratumn/sdk/tendermint"
	"github.com/stratumn/sdk/tmpop"
)

var (
	path    = flag.String("path", filestore.DefaultPath, "path to directory where files are stored")
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()

	a := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})
	adapterInfo, err := a.GetInfo()
	if err != nil {
		log.Fatal(err)
	}

	tmpopConfig := &tmpop.Config{Commit: commit, Version: version, DbDir: *path}
	tmpop := tmpop.New(a, tmpopConfig)

	log.Infof("TMPop v%s@%s", tmpopConfig.Version, tmpopConfig.Commit[:7])
	log.Infof("Adapter %v", adapterInfo)
	log.Info("Copyright (c) 2017 Stratumn SAS")
	log.Info("Mozilla Public License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	tendermint.RunNodeForever(tendermint.GetConfig(), tmpop)
}
