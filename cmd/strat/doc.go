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

// The command start contains various subcommands related to Stratumn.
//
// Usage
//
// 	The Stratumn CLI provides various commands to generate and work with Stratumn's technology.
//
// 	Usage:
// 	  strat [command]
//
// 	Available Commands:
// 	  build       Build project
// 	  deploy      Deploy project
// 	  down        Stop project services
// 	  generate    Generate a project
// 	  generators  List available generators
// 	  info        Display Stratumn CLI info
// 	  pull        Pull project Docker images
// 	  push        Push project Docker images
// 	  run         Run project command
// 	  test        Run project test suite
// 	  up          Start project services
// 	  update      Update Stratumn CLI or generators
// 	  version     Display version info
//
// 	Flags:
// 	  -c, --config string             Location of Stratumn configuration files (default "/Users/stephan/.stratumn")
// 	      --generators-owner string   Github owner of generators repository (default "stratumn")
// 	  -p, --generators-path string    Location where generators are stored locally (default "/Users/stephan/.stratumn/generators")
// 	      --generators-ref string     Git branch, tag, or commit of generators repository (default "master")
// 	      --generators-repo string    Name of generators Git repository (default "generators")
// 	      --gh-token string           Github API token
// 	      --stdin                     Attach stdin to process when executing project commands (default true)
//
// 	Use "strat [command] --help" for more information about a command.
package main
