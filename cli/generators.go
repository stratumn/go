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
	"os"
	"text/tabwriter"

	"github.com/google/subcommands"
	"github.com/stratumn/go/generator/repo"
	"golang.org/x/net/context"
)

// Generators is a command that lists generators.
type Generators struct {
	owner string
	repo  string
	ref   string
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Generators) Name() string {
	return "generators"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Generators) Synopsis() string {
	return "list generators"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Generators) Usage() string {
	return `generators:
  List generators.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Generators) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.owner, "owner", DefaultGeneratorsOwner, "Github owner")
	f.StringVar(&cmd.repo, "repo", DefaultGeneratorsRepo, "Github repository")
	f.StringVar(&cmd.ref, "ref", DefaultGeneratorsRef, "Github branch, tag, or commit SHA1")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Generators) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
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

	list, err := repo.List(cmd.ref)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, "NAME\tDESCRIPTION\tAUTHOR\tVERSION\tLICENSE"); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	for _, desc := range list {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			desc.Name, desc.Description, desc.Author, desc.Version, desc.License); err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
	}

	if err := tw.Flush(); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
