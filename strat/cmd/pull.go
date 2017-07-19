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

import "github.com/spf13/cobra"

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull [args...]",
	Short: "Pull project Docker images",
	Long: `Pull Docker images of project in current directory.

It executes, if present, the pull command of the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScript(PullScript, "", args, false, useStdin)
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
