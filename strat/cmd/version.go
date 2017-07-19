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

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version info",
	Long: `Display the version info of Stratumn CLI.

It outputs the semver string, and the first seven characters of the Git hash.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("unexpected arguments")
		}
		fmt.Printf("%s v%s@%s\n", "Stratumn CLI", version, commit[:7])
		return nil
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
