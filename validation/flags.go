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

package validation

import (
	"flag"
)

var (
	rulesPath   string
	pluginsPath string
)

// RegisterFlags registers the command-line monitoring flags.
func RegisterFlags() {
	flag.StringVar(&rulesPath, "rules_path", DefaultFilename, "Path to the file containing validation rules")
	flag.StringVar(&pluginsPath, "plugins_path", DefaultPluginsDirectory, "Path to the directory containing validation plugins")
}

// ConfigurationFromFlags builds configuration from user-provided
// command-line flags.
func ConfigurationFromFlags() *Config {
	return &Config{
		RulesPath:   rulesPath,
		PluginsPath: pluginsPath,
	}
}
