// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/filestore"
	"github.com/stratumn/sdk/tmpop"
)

var (
	path      = flag.String("path", filestore.DefaultPath, "path to directory where files are stored")
	cacheSize = flag.Int("cacheSize", tmpop.DefaultCacheSize, "size of the cache of the storage tree")
	version   = "0.1.0"
	commit    = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()

	a, err := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})
	if err != nil {
		log.Fatal(err)
	}

	tmpopConfig := &tmpop.Config{Commit: commit, Version: version, CacheSize: *cacheSize}

	tmpop.Run(a, tmpopConfig)
}
