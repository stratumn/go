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
	"golang.org/x/net/context"
)

// Version is a command that prints the version.
type Version struct {
	Version string
	Commit  string
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Version) Name() string {
	return "version"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Version) Synopsis() string {
	return "print version info"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Version) Usage() string {
	return `version:
  Print version info.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Version) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Version) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	fmt.Printf("%s v%s@%s\n", "Stratumn CLI", cmd.Version, cmd.Commit[:7])

	return subcommands.ExitSuccess
}
