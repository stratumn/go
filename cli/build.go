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

// Build is a project command that runs the build script.
type Build struct {
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Build) Name() string {
	return "build"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Build) Synopsis() string {
	return "run build script"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Build) Usage() string {
	return `build:
  Run build script.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Build) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Build) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	return runScript(BuildScript, "", false)
}
