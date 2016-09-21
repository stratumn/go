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

// Package cli implements command line commands.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/google/subcommands"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	// DefaultGeneratorsOwner is the  default owner of the generators' Github repository.
	DefaultGeneratorsOwner = "stratumn"

	// DefaultGeneratorsRepo is the default name of the generators' Github repository.
	DefaultGeneratorsRepo = "generators"

	// StratumnDir is the name of the Stratumn directory within the home folder.
	StratumnDir = ".stratumn"

	// GeneratorsDir is the name of the generators directory within StratumnDir.
	GeneratorsDir = "generators"

	// VarsFile is the name of the variable file within StratumnDir.
	VarsFile = "variables.json"

	// ProjectFile is the name of the project file within the project directory.
	ProjectFile = "stratumn.json"

	// UpScript is the name of the project serve script.
	UpScript = "up"

	// TestScript is the name of the project test script.
	TestScript = "test"
)

// Project describes a project.
type Project struct {
	Scripts map[string]string `json:"scripts"`
}

// NewProjectFromFile instantiates a project from a project file.
func NewProjectFromFile(path string) (*Project, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(f)
	var prj Project
	if err := dec.Decode(&prj); err != nil {
		return nil, err
	}
	return &prj, nil
}

// GetScript returns a script by name.
// If the script is undefined, it returns an empty string.
func (prj *Project) GetScript(name string) string {
	v, ok := prj.Scripts[name]
	if ok {
		return v
	}
	return ""
}

func generatorPath(owner, repo string) (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, StratumnDir, GeneratorsDir, owner, repo), nil
}

func varsPath() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, StratumnDir, VarsFile), nil
}

func runScript(name string) subcommands.ExitStatus {
	prj, err := NewProjectFromFile(ProjectFile)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	script := prj.GetScript(name)
	if script == "" {
		fmt.Printf("Project doesn't have a %q script.\n", name)
		return subcommands.ExitFailure
	}

	parts := strings.Split(script, " ")
	c := exec.Command(parts[0], parts[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Start(); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		for range sigc {
		}
	}()

	if err := c.Wait(); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
