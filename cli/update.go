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

// Update is a command to update the CLI.
type Update struct {
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Update) Name() string {
	return "update"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Update) Synopsis() string {
	return "Update the CLI"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Update) Usage() string {
	return `update:
  Update the CLI.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Update) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Update) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	path, err := generatorPath()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	repo := repo.New(path, GeneratorsOwner, GeneratorsRepo)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	fmt.Println("Updating generators...")

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
