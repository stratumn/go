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
	"runtime"

	"github.com/spf13/cobra"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display Stratumn CLI info",
	Long: `Display information about Stratumn CLI.

It outputs version, copyright, license, and runtime information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("unexpected arguments")
		}
		fmt.Printf("%s v%s@%s\n", "Stratumn CLI", version, commit[:7])
		fmt.Println("Copyright (c) 2017 Stratumn SAS")
		fmt.Println("Apache License 2.0")
		fmt.Printf("Runtime %s %s %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}
