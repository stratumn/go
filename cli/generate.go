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
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/subcommands"
	"github.com/stratumn/go/generator"
	"github.com/stratumn/go/generator/repo"
	"golang.org/x/net/context"
)

// Generate is a command to generate projects.
type Generate struct {
	repo      string
	generator string
	owner     string
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Generate) Name() string {
	return "generate"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Generate) Synopsis() string {
	return "generate a project"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Generate) Usage() string {
	return `generate [flags] [out]:
  Generate a project.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Generate) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.owner, "owner", "", "Github owner")
	f.StringVar(&cmd.repo, "repo", "", "Github repository")
	f.StringVar(&cmd.generator, "generator", "", "generator name")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Generate) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 1 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	out := args[0]

	if cmd.owner == "" {
		cmd.owner = DefaultGeneratorsOwner
	}
	if cmd.repo == "" {
		cmd.repo = DefaultGeneratorsRepo
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

	name := cmd.generator
	if name == "" {
		list, err := repo.List()
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		in := generator.StringSelect{
			InputShared: generator.InputShared{
				Prompt: "What would you like to generate?",
			},
			Options: []generator.StringSelectOption{},
		}
		for i, desc := range list {
			in.Options = append(in.Options, generator.StringSelectOption{
				Input: strconv.Itoa(i + 1),
				Value: desc.Name,
				Text:  desc.Description,
			})
		}

		fmt.Print(in.Msg())
		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("? ")
			str, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return subcommands.ExitFailure
			}
			str = strings.TrimSpace(str)
			if err := in.Set(str); err != nil {
				fmt.Println(err)
				continue
			}
			name = in.Get().(string)
			break
		}
	}

	varsPath, err := varsPath()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	varsFile, err := os.Open(varsPath)

	vars := map[string]interface{}{}

	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
	} else {
		dec := json.NewDecoder(varsFile)
		if err := dec.Decode(&vars); err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
	}

	vars["dir"] = filepath.Base(out)

	opts := generator.Options{
		DefVars:  vars,
		TmplVars: vars,
	}

	if err := repo.Generate(name, out, &opts); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	fmt.Println("Done!")

	return subcommands.ExitSuccess
}
