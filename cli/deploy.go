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

// Deploy is a project command that deploys a project to an environment.
type Deploy struct {
	stdin bool
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Deploy) Name() string {
	return "deploy"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Deploy) Synopsis() string {
	return "deploy project to an environment"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Deploy) Usage() string {
	return `deploy environment [args...]:
  Deploy project to given environment.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Deploy) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.stdin, "stdin", false, "attach stdin to command")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Deploy) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) < 1 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	script := fmt.Sprintf(DeployScriptFmt, args[0])

	return runScript(script, "", args[1:], false, cmd.stdin)
}
