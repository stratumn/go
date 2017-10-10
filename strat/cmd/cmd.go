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

// Package cmd implements command line commands for strat.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	// EnvPrefix is the prefix of environment variables to set flag values.
	EnvPrefix = "stratumn"

	// Owner is the owner of the CLI's Github repository.
	Owner = "stratumn"

	// Repo is the name of the CLI's Github repository.
	Repo = "sdk"

	// AssetFormat is the format of the CLI GIthub asset.
	AssetFormat = "strat-%s-%s.zip"

	// AssetBinary is the file name of the binary within the CLI asset.
	AssetBinary = "strat/strat"

	// AssetBinaryWin is the file name of the binary within the CLI asset
	// on Windows.
	AssetBinaryWin = "strat/strat.exe"

	// OldBinary is the name of the old binary after an update.
	OldBinary = ".strat.old"

	// SigExt the extension of the signature of the binary.
	SigExt = ".sig"

	// DefaultGeneratorsOwner is the default owner of the generators' Github
	// repository.
	DefaultGeneratorsOwner = "stratumn"

	// DefaultGeneratorsRepo is the default name of the generators' Github
	// repository.
	DefaultGeneratorsRepo = "generators"

	// StratumnConfigEnv is the name of the environment variable to override
	// the default configuration path.
	StratumnConfigEnv = "STRATUMN_CONFIG"

	// DefaultStratumnDir is the name of the Stratumn directory within the
	// home folder.
	DefaultStratumnDir = ".stratumn"

	// GeneratorsDir is the name of the generators directory within the
	// configuration directory.
	GeneratorsDir = "generators"

	// VarsFile is the name of the variable file within the configuration
	// directory.
	VarsFile = "variables.json"

	// ProjectFile is the name of the project file within the project
	// directory.
	ProjectFile = "stratumn.json"

	// InitScript is the name of the project init script.
	InitScript = "init"

	// UpScript is the name of the project up script.
	UpScript = "up"

	// DownScript is the name of the project down script.
	DownScript = "down"

	// BuildScript is the name of the project build script.
	BuildScript = "build"

	// TestScript is the name of the project test script.
	TestScript = "test"

	// PullScript is the name of the project pull script.
	PullScript = "pull"

	// PushScript is the name of the project push script.
	PushScript = "push"

	// DeployScriptFmt is the format of the name of the project deploy
	// script for an environment.
	DeployScriptFmt = "deploy:%s"
)

var (
	// DefaultGeneratorsRef is the default reference of the generators'
	// Github repository. It is a variable because it is overridden at
	// compile time.
	DefaultGeneratorsRef = "master"
)

const (
	win    = "windows"
	darwin = "darwin"
)

var (
	nixShell    = []string{"sh", "-i", "-c"}
	winShell    = []string{"cmd", "/C"}
	brewInfoCmd = []string{"brew", "info", "--json=v1", "stratumn/sdk/strat"}
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

func generatorPath() string {
	if generatorsUseLocalFiles {
		return generatorsPath
	}
	return filepath.Join(generatorsPath, generatorsOwner, generatorsRepo)
}

func varsPath() string {
	return filepath.Join(cfgPath, VarsFile)
}

func runScript(name, wd string, args []string, ignoreNotExist bool) error {
	if wd != "" {
		if err := os.Chdir(wd); err != nil {
			return err
		}
	}

	prj, err := NewProjectFromFile(ProjectFile)
	if err != nil {
		return err
	}

	// Look for OS/Arch specific scripts first.
	script := prj.GetScript(name + ":" + runtime.GOOS + ":" + runtime.GOARCH)
	if script == "" {
		script = prj.GetScript(name + ":" + runtime.GOOS)
	}
	if script == "" {
		script = prj.GetScript(name + ":" + runtime.GOARCH)
	}
	if script == "" {
		script = prj.GetScript(name)
	}
	if script == "" {
		if ignoreNotExist {
			return nil
		}
		return fmt.Errorf("project doesn't have a %q script", name)
	}

	if len(args) > 0 {
		script += " " + strings.Join(args, " ")
	}

	var shell []string
	if runtime.GOOS == win {
		shell = winShell
	} else {
		shell = nixShell
	}

	// Look for full path of shell binary.
	bin, err := exec.LookPath(shell[0])
	if err != nil {
		return err
	}

	argv := append(shell, script)

	// Calling syscall.Exec replaces the current process with the command.
	return syscall.Exec(bin, argv, os.Environ())
}
