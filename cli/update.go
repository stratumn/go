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

package cli

import (
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/stratumn/go/generator/repo"
	"golang.org/x/net/context"
)

// Update is a command that updates the CLI.
type Update struct {
	owner string
	repo  string
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Update) Name() string {
	return "update"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Update) Synopsis() string {
	return "update the CLI"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Update) Usage() string {
	return `update:
  Update the CLI.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Update) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.owner, "owner", "", "Github owner")
	f.StringVar(&cmd.repo, "repo", "", "Github repository")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Update) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	fmt.Println("Updating generators...")

	if cmd.owner == "" {
		cmd.owner = DefaultGeneratorsOwner
	}
	if cmd.repo == "" {
		cmd.repo = DefaultGeneratorsRepo
	}

	path, err := generatorPath(cmd.owner, cmd.repo)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	repo := repo.New(path, cmd.owner, cmd.repo)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	_, updated, err := repo.Update()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if updated {
		fmt.Println("Generators updated successfully.")
	} else {
		fmt.Println("Generators are already up-to-date.")
	}

	return subcommands.ExitSuccess
}
