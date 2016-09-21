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
	"runtime"

	"github.com/google/subcommands"
	"golang.org/x/net/context"
)

// Info is a command that prints info about the program.
type Info struct {
	Version string
	Commit  string
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Info) Name() string {
	return "info"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Info) Synopsis() string {
	return "print program info"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Info) Usage() string {
	return `info:
  Print program info.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (*Info) SetFlags(f *flag.FlagSet) {
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Info) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	fmt.Printf("%s v%s@%s\n", "Stratumn CLI", cmd.Version, cmd.Commit[:7])
	fmt.Print("Copyright (c) 2016 Stratumn SAS")
	fmt.Print("Apache License 2.0")
	fmt.Printf("Runtime %s %s %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	return subcommands.ExitSuccess
}
