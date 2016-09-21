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
	"os"

	"golang.org/x/net/context"

	"github.com/google/subcommands"
	"github.com/stratumn/go/cli"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.FlagsCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")
	subcommands.Register(&cli.Generators{}, "projects")
	subcommands.Register(&cli.Generate{}, "projects")
	subcommands.Register(&cli.Up{}, "projects")
	subcommands.Register(&cli.Test{}, "projects")
	subcommands.Register(&cli.Update{}, "cli")
	subcommands.Register(&cli.Version{Version: version, Commit: commit}, "cli")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
