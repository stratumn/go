// Copyright 2017 Stratumn SAS. All rights reserved.
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

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stratumn/go-indigocore/generator"
	"github.com/stratumn/go-indigocore/generator/repo"
)

var (
	generatorName string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Aliases: []string{"g"},
	Use:     "generate <path>",
	Short:   "Generate a project",
	Long: `Generate a project using a generator.

It asks which generator to use, then uses that generator to generate a project in the given path.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("expected path")
		}
		if len(args) > 1 {
			return errors.New("unexpected argument")
		}

		out := args[0]

		path := generatorPath()
		repo := repo.New(path, generatorsOwner, generatorsRepo, ghToken, !generatorsUseLocalFiles)

		name := generatorName
		if name == "" {
			list, err := repo.List(generatorsRef)
			if err != nil {
				return err
			}

			in := generator.StringSelect{
				InputShared: generator.InputShared{
					Prompt: "What would you like to generate?",
				},
				Options: generator.StringSelectOptions{},
			}
			for _, desc := range list {
				in.Options[desc.Name] = desc.Description
			}
			ret, err := in.Run()
			name = ret.(string)
		}

		varsFile, err := os.Open(varsPath())
		vars := map[string]interface{}{}
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			dec := json.NewDecoder(varsFile)
			if err := dec.Decode(&vars); err != nil {
				return err
			}
		}

		vars["dir"] = filepath.Base(out)

		opts := generator.Options{
			DefVars:  vars,
			TmplVars: vars,
		}

		if err := repo.Generate(name, out, &opts, generatorsRef); err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Join(out, ProjectFile)); err == nil {
			if err = runScript(InitScript, out, nil, true); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}

		fmt.Println("Done!")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVarP(
		&generatorName,
		"generator-name",
		"n",
		"",
		"Specify generator name instead of asking",
	)
}
