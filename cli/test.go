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

// Test is a project command that runs tests.
type Test struct {
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Test) Name() string {
	return "test"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Test) Synopsis() string {
	return "run tests"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Test) Usage() string {
	return `test [args...]:
  Run tests.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Test) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Test) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	testRes := runScript(TestScript, "", f.Args(), false)
	downRes := runScript(DownTestScript, "", nil, true)

	if testRes != subcommands.ExitSuccess {
		return testRes
	}

	return downRes
}
