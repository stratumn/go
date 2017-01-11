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

package main

import (
	"flag"

	"github.com/stratumn/go/filestore"
	"github.com/stratumn/go/tmpop"
)

var (
	addrPtr = flag.String("addr", "tcp://0.0.0.0:46658", "Listen address")
	abciPtr = flag.String("tmsp", "socket", "TMSP server: socket | grpc")
	path    = flag.String("path", filestore.DefaultPath, "path to directory where files are stored")
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func main() {
	flag.Parse()

	a := filestore.New(&filestore.Config{Path: *path, Version: version, Commit: commit})

	tmpop.Run(a, &tmpop.Config{Commit: commit, Version: version, DbDir: *path}, addrPtr, abciPtr)
}
