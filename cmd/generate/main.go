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
	"log"
	"os"
	"runtime"

	"github.com/stratumn/go/generator"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func main() {
	log.SetFlags(0)
	log.Printf("%s v%s@%s", "Stratumn Generate", version, commit[:7])
	log.Print("Copyright (c) 2016 Stratumn SAS")
	log.Print("Apache License 2.0")
	log.Printf("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatalf("usage: %s src dst", os.Args[0])
	}
	src, dst := args[0], args[1]
	gen, err := generator.NewFromDir(src, &generator.Options{})
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	if err := gen.Exec(dst); err != nil {
		log.Fatalf("err: %s", err)
	}

	log.Printf("Done")
}
