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

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test [args...]",
	Short: "Run project test suite",
	Long: `Run the test suite of project in current directory.

It executes, if present, the test command of the project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		testErr := runScript(TestScript, "", args, false, useStdin)
		downErr := runScript(DownTestScript, "", nil, true, useStdin)

		if testErr != nil {
			return testErr
		}

		return downErr
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
