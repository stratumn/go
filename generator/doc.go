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

// Package generator deals with creating projects from template files.
//
// A generator is a directory containing a definition file, template files, and
// partials.
//
// In addition to metadata, the definition file, which must be a JSON document
// named `generator.json` at the root of the generator, can define variables and
// user inputs that will be made available to the templates and partials. An
// input is read from the user only when its value is needed for the first time.
//
// The templates are files that should be placed in a `files` directory, and use
// the Go template syntax to produce project files. Every template will result
// in a file of the same name being created in the generated project directory.
// The path of the generated file within the generated project wil be the path
// of the template relative to the `files` directory.
//
// Templates can include partials via the `partial` function. The partials
// should be placed in a `partials` directory. As opposed to templates, partials
// will not result in files being generated. The `partial` function expects the
// path of a partial relative to the `partials` directory, and optionally a
// variadic list of variable maps for the partial. The partials have access to
// the same functions and variables as the templates.
//
// By default templates are evaluated in alphabetical order. You can have more
// control over the order by adding a `priorities` array to the definition file.
// This array should contain a list of files relative to the `files` directory
// that will be evaluated first. That way it is possible to control the order
// which inputs will be read from the user.
//
// A basic definition file may look something like this:
//      {
//        "name": "basic",
//        "version": "0.1.0",
//        "description": "A basic generator",
//        "author": "Stratumn",
//        "license": "MIT",
//        "inputs": {
//          "name": {
//            "type": "string",
//            "prompt": "Project name:",
//            "default": "{{.dir}}",
//            "format": ".+"
//          }
//        }
//      }
//
// In this case, one input called `name` of type `string` is defined. Its
// default value is `{{.dir}}`, which should be a variable given to the
// definition file parser.
//
// A template file in the `template` directory can access the user input for
// `name` using the template function `input`. For instance it could be a
// Markdown file containing the following:
//      # {{input "name"}}
//      A basic project
//
// A project can be generated from the generator this way:
//      // Directory where the project will be generated.
//      dst := "path/to/generated/project"
//
//      // Add a `dir` variable for the definition file set to the name
//      // of the project directory.
//      opts := generator.Options{
//              DefVars: map[string]interface{}{
//                      "dir": filepath.Dir(dst),
//              },
//      }
//
//      // Load the generator.
//      gen, err := generator.NewFromDir("path/to/generator", &opts)
//      if err != nil {
//              panic(err)
//      }
//
//      // Generate the project.
//      if err := gen.Exec(dst); err != nil {
//              panic(err)
//      }
package generator
