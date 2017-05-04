// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command rethinktmpop starts a tmpop node with a rethinkstore.
package main

import (
	"flag"

	"github.com/stratumn/goprivate/rethinkstore"
	"github.com/stratumn/sdk/tmpop"
)

var (
	cacheSize = flag.Int("cacheSize", tmpop.DefaultCacheSize, "size of the cache of the storage tree")
	version   = "0.1.0"
	commit    = "00000000000000000000000000000000"
)

func init() {
	rethinkstore.RegisterFlags()
}

func main() {
	flag.Parse()

	a := rethinkstore.InitializeWithFlags(version, commit)

	tmpopConfig := &tmpop.Config{Commit: commit, Version: version, CacheSize: *cacheSize}

	tmpop.Run(a, tmpopConfig)
}
