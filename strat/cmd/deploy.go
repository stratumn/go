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

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy <env> [args...]",
	Short: "Deploy project",
	Long: `Deploy project in current directory to an environment.

It executes, if present, the deploy:<env> command of the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("expected environment")
		}
		script := fmt.Sprintf(DeployScriptFmt, args[0])
		return runScript(script, "", args[1:], false, useStdin)
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
