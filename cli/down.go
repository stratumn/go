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

// Down is a project command that stops the services.
type Down struct {
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Down) Name() string {
	return "down"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Down) Synopsis() string {
	return "stop services"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Down) Usage() string {
	return `down [args...]:
  Stop services.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Down) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Down) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return runScript(DownScript, "", f.Args(), false)
}
