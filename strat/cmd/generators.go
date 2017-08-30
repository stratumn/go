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
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stratumn/sdk/generator/repo"
)

// generatorsCmd represents the generators command
var generatorsCmd = &cobra.Command{
	Use:   "generators",
	Short: "List available generators",
	Long: `List the available generators.
	
It outputs the name, description, author, version, and license of each generator.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("unexpected arguments")
		}

		path := generatorPath()
		repo := repo.New(path, generatorsOwner, generatorsRepo, ghToken, !generatorsUseLocalFiles)
		list, err := repo.List(generatorsRef)
		if err != nil {
			return err
		}

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if _, err := fmt.Fprintln(tw, "NAME\tDESCRIPTION\tAUTHOR\tVERSION\tLICENSE"); err != nil {
			return err
		}

		for _, desc := range list {
			if _, err := fmt.Fprintf(
				tw,
				"%s\t%s\t%s\t%s\t%s\n",
				desc.Name,
				desc.Description,
				desc.Author,
				desc.Version,
				desc.License,
			); err != nil {
				return err
			}
		}

		return tw.Flush()
	},
}

func init() {
	RootCmd.AddCommand(generatorsCmd)
}
