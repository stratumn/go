// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command filestore starts a storehttp server with a filestore.
package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/filestore"
	"github.com/stratumn/sdk/store/storehttp"
)

var (
	path    = flag.String("path", filestore.DefaultPath, "Path to directory where files are stored")
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func init() {
	storehttp.RegisterFlags()
}

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", filestore.Description, version, commit[:7])
	a, err := filestore.New(&filestore.Config{
		Path:    *path,
		Version: version,
		Commit:  commit,
	})
	if err != nil {
		log.Fatal(err)
	}
	storehttp.RunWithFlags(a)
}
