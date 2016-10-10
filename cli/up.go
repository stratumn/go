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

	"github.com/google/subcommands"
	"golang.org/x/net/context"
)

// Up is a project command that starts the services.
type Up struct {
	stdin bool
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Up) Name() string {
	return "up"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Up) Synopsis() string {
	return "start services"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Up) Usage() string {
	return `up [args...]:
  Start services.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Up) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.stdin, "stdin", false, "attach stdin to command")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Up) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return runScript(UpScript, "", f.Args(), false, cmd.stdin)
}
