// Copyright 2016 Stratumn SAS. All rights reserved.
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

package main

import (
	"flag"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/go/filestore"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store/storehttp"
)

var (
	http     = flag.String("http", storehttp.DefaultAddress, "HTTP address")
	path     = flag.String("path", filestore.DefaultPath, "path to directory where files are stored")
	certFile = flag.String("tlscert", "", "TLS certificate file")
	keyFile  = flag.String("tlskey", "", "TLS private key file")
	version  = "0.1.0"
	commit   = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()
	log.Infof("%s v%s@%s", filestore.Description, version, commit[:7])

	a := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})

	c := &jsonhttp.Config{
		Address:  *http,
		CertFile: *certFile,
		KeyFile:  *keyFile,
	}
	storehttp.Run(a, c)
}
