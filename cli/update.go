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
	"path/filepath"
	"strings"

	"github.com/google/subcommands"
	"github.com/stratumn/go/generator/repo"
	"golang.org/x/net/context"
)

// Update is a command that updates the CLI.
type Update struct {
	prerelease bool
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
	f.BoolVar(&cmd.prerelease, "prerelease", false, "update to prerelease if available")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Update) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	fmt.Println("Updating generators...")

	path, err := generatorsPath()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	matches, err := filepath.Glob(filepath.Join(path, "*", "*", repo.StatesDir, "*"))
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	for _, path = range matches {
		var (
			parts = strings.Split(path, string(filepath.Separator))
			l     = len(parts)
			owner = parts[l-4]
			rep   = parts[l-3]
			ref   = parts[l-1]
			name  = fmt.Sprintf("%s/%s@%s", owner, rep, ref)
			p     = filepath.Join(parts[:l-2]...)
		)

		fmt.Printf("Updating generators %q...\n", name)

		r := repo.New(p, owner, rep)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		_, updated, err := r.Update(ref)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		if updated {
			fmt.Printf("Generators %q updated successfully.\n", name)
		} else {
			fmt.Printf("Generators %q already up-to-date.\n", name)
		}
	}

	fmt.Println("Generators updated successfully.")

	return subcommands.ExitSuccess
}
