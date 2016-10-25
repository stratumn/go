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

package cmd

import "github.com/spf13/cobra"

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down [args...]",
	Short: "Stop project services",
	Long: `Stop services started by project in current directory.

It executes, if present, the down command of the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScript(DownScript, "", args, false, useStdin)
	},
}

func init() {
	RootCmd.AddCommand(downCmd)
}
