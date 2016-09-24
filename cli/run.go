// Copyright 2016 Stratumn SAS. All rights rerund.
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

// Run is a project command that runs script by name.
type Run struct {
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Run) Name() string {
	return "run"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Run) Synopsis() string {
	return "run script by name"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Run) Usage() string {
	return `run script:
  Run script by name.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Run) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Run) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 1 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	return runScript(args[0], "", false)
}
