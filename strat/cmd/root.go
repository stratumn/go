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
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

var (
	cfgPath                 string
	ghToken                 string
	generatorsPath          string
	generatorsOwner         string
	generatorsRepo          string
	generatorsRef           string
	generatorsUseLocalFiles bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "strat",
	Short: "Stratumn CLI",
	Long:  `The Stratumn CLI provides various commands to generate and work with Stratumn's technology.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ver, com string) {
	version = ver
	commit = com
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	defCfgPath := filepath.Join(homeDir, DefaultStratumnDir)
	defGeneratorsPath := filepath.Join(defCfgPath, GeneratorsDir)

	RootCmd.PersistentFlags().StringVarP(
		&cfgPath,
		"config",
		"c",
		defCfgPath,
		"Location of Stratumn configuration files",
	)

	RootCmd.PersistentFlags().StringVar(
		&ghToken,
		"github-api-token",
		"",
		"Github API token for private repositories",
	)

	RootCmd.PersistentFlags().StringVarP(
		&generatorsPath,
		"generators-path",
		"p",
		defGeneratorsPath,
		"Location where generators are stored locally",
	)

	RootCmd.PersistentFlags().StringVar(
		&generatorsOwner,
		"generators-owner",
		DefaultGeneratorsOwner,
		"Github owner of generators repository",
	)

	RootCmd.PersistentFlags().StringVar(
		&generatorsRepo,
		"generators-repo",
		DefaultGeneratorsRepo,
		"Name of generators Git repository",
	)

	RootCmd.PersistentFlags().StringVar(
		&generatorsRef,
		"generators-ref",
		DefaultGeneratorsRef,
		"Git branch, tag, or commit of generators repository",
	)

	RootCmd.PersistentFlags().BoolVar(
		&generatorsUseLocalFiles,
		"generators-use-local-files",
		false,
		"Do not retrieve generators files from git",
	)
}

// initConfig reads ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv() // read in environment variables that match
}
