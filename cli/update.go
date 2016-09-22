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

package cli

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/github"
	"github.com/google/subcommands"
	"github.com/kardianos/osext"
	"github.com/stratumn/go/generator/repo"
	"golang.org/x/net/context"
)

// Update is a command that updates the CLI.
type Update struct {
	Version    string
	generators bool
	cli        bool
	prerelease bool
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Update) Name() string {
	return "update"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Update) Synopsis() string {
	return "update the CLI"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Update) Usage() string {
	return `update:
  Update the CLI.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Update) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.generators, "generators", true, "update generators")
	f.BoolVar(&cmd.cli, "cli", true, "update CLI")
	f.BoolVar(&cmd.prerelease, "prerelease", false, "update to prerelease")
}

// Execute implements github.com/google/subcommands.Command.Execute().
func (cmd *Update) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) > 0 {
		fmt.Println(cmd.Usage())
		return subcommands.ExitUsageError
	}

	if cmd.generators {
		if code := cmd.updateGenerators(); code != subcommands.ExitSuccess {
			return code
		}
	}

	if cmd.cli {
		if code := cmd.updateCLI(); code != subcommands.ExitSuccess {
			return code
		}
	}

	return subcommands.ExitSuccess
}

func (cmd *Update) updateGenerators() subcommands.ExitStatus {
	fmt.Println("Updating generators...")

	path, err := generatorsPath()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	matches, err := filepath.Glob(filepath.Join(path, "*", "*", repo.StatesDir, "*"))
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	for _, path = range matches {
		var (
			parts = strings.Split(path, string(filepath.Separator))
			l     = len(parts)
			owner = parts[l-4]
			rep   = parts[l-3]
			ref   = parts[l-1]
			name  = fmt.Sprintf("%s/%s@%s", owner, rep, ref)
			p     = filepath.Join(parts[:l-2]...)
		)

		fmt.Printf("Updating generators %q...\n", name)

		r := repo.New(p, owner, rep)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		_, updated, err := r.Update(ref)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		if updated {
			fmt.Printf("Generators %q updated successfully.\n", name)
		} else {
			fmt.Printf("Generators %q already up-to-date.\n", name)
		}
	}

	fmt.Println("Generators updated successfully.")

	return subcommands.ExitSuccess
}

func (cmd *Update) updateCLI() subcommands.ExitStatus {
	fmt.Println("Updating CLI...")

	client := github.NewClient(nil)
	rels, res, err := client.Repositories.ListReleases(CLIOwner, CLIRepo, nil)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer res.Body.Close()

	var (
		asset *github.ReleaseAsset
		tag   *string
	)
	for _, r := range rels {
		if *r.Prerelease == cmd.prerelease {
			if *r.TagName != "v"+cmd.Version {
				name := fmt.Sprintf(CLIAssetFormat, runtime.GOOS, runtime.GOARCH)
				for _, a := range r.Assets {
					if *a.Name == name {
						asset = &a
						tag = r.TagName
						break
					}
				}
			}
			break
		}
	}

	if asset == nil {
		fmt.Println("CLI already up-to-date.")
		return subcommands.ExitSuccess
	}

	fmt.Printf("Found new version %q.\n", *tag)

	execPath, err := osext.Executable()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	name := filepath.Base(execPath)

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer os.RemoveAll(tempDir)

	tempZipFile := filepath.Join(tempDir, name+".zip")
	f, err := os.OpenFile(tempZipFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer f.Close()

	fmt.Printf("Downloading %q...\n", *asset.Name)

	rc, url, err := client.Repositories.DownloadReleaseAsset(CLIOwner, CLIRepo, *asset.ID)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	var r io.ReadCloser

	if rc != nil {
		r = rc
	} else if url != "" {
		res, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}
		r = res.Body
	}
	defer r.Close()

	if _, err := io.Copy(f, r); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	zr, err := zip.OpenReader(tempZipFile)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Extracting %q...\n", *asset.Name)

	for _, f := range zr.File {
		if f.Name == CLIAssetBinary {
			info, err := os.Stat(execPath)
			if err != nil {
				fmt.Println(err)
				return subcommands.ExitFailure
			}

			rc, err := f.Open()
			if err != nil {
				fmt.Println(err)
				return subcommands.ExitFailure
			}
			defer rc.Close()

			dst, err := os.OpenFile(execPath, os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				fmt.Println(err)
				return subcommands.ExitFailure
			}
			defer dst.Close()

			if _, err := io.Copy(dst, rc); err != nil {
				fmt.Println(err)
				return subcommands.ExitFailure
			}

			fmt.Println("CLI updated successfully.")
			return subcommands.ExitSuccess
		}
	}

	fmt.Printf("Could not find %s in %s.\n", CLIAssetBinary, *asset.Name)
	return subcommands.ExitFailure
}
