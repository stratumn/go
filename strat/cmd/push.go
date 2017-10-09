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

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push [args...]",
	Short: "Push project Docker images",
	Long: `Push the Docker images of project in current directory.

It executes, if present, the push command of the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScript(PushScript, "", args, false)
	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
}
