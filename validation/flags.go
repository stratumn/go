// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package validation

import (
	"flag"
)

const (
	// Component name for monitoring.
	Component = "validation"

	// DefaultFilename is the default filename for the file containing
	// validation rules.
	// This is used in non-blockchain scenario.
	DefaultFilename = "/data/validation/rules.json"

	// DefaultPluginsDirectory is the default directory where validation
	// plugins should be stored.
	DefaultPluginsDirectory = "/data/validation/plugins"
)

var (
	rulesPath   string
	pluginsPath string
)

// Config contains the path of the rules JSON file and the directory where the validator scripts are located.
type Config struct {
	RulesPath   string
	PluginsPath string
}

// RegisterFlags registers the command-line monitoring flags.
func RegisterFlags() {
	flag.StringVar(&rulesPath, "rules_path", "", "Path to the file containing validation rules")
	flag.StringVar(&pluginsPath, "plugins_path", "", "Path to the directory containing validation plugins")
}

// ConfigurationFromFlags builds configuration from user-provided command-line
// flags.
func ConfigurationFromFlags() *Config {
	return &Config{
		RulesPath:   rulesPath,
		PluginsPath: pluginsPath,
	}
}
