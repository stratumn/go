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
	"fmt"
	"os"

	"golang.org/x/net/context"

	"github.com/google/subcommands"
	"github.com/kardianos/osext"
	"github.com/stratumn/go/cli"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func main() {
	// Remove old binary if found.
	if execPath, err := osext.Executable(); err == nil {
		path := execPath + cli.CLIOldExt
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove old executable %q.\n", path)
			fmt.Println("You can remove it manualy.")
			fmt.Printf("The error was: %q.\n", err)
		}
	} else {
		fmt.Printf("Warning: %s.\n", err)
	}

	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.FlagsCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")
	subcommands.Register(&cli.Generators{}, "generator")
	subcommands.Register(&cli.Generate{}, "generator")
	subcommands.Register(&cli.Build{}, "project")
	subcommands.Register(&cli.Up{}, "project")
	subcommands.Register(&cli.Test{}, "project")
	subcommands.Register(&cli.Run{}, "project")
	subcommands.Register(&cli.Update{Version: version}, "CLI")
	subcommands.Register(&cli.Info{Version: version, Commit: commit}, "CLI")
	subcommands.Register(&cli.Version{Version: version, Commit: commit}, "CLI")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
