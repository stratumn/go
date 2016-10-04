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
	"bytes"
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
	"golang.org/x/crypto/openpgp"
	"golang.org/x/net/context"
)

// Update is a CLI command that updates the CLI or generators.
type Update struct {
	Version    string
	generators bool
	prerelease bool
	force      bool
	ghToken    string
}

// Name implements github.com/google/subcommands.Command.Name().
func (*Update) Name() string {
	return "update"
}

// Synopsis implements github.com/google/subcommands.Command.Synopsis().
func (*Update) Synopsis() string {
	return "update the CLI or generators"
}

// Usage implements github.com/google/subcommands.Command.Usage().
func (*Update) Usage() string {
	return `update:
  Update the CLI or generators.
`
}

// SetFlags implements github.com/google/subcommands.Command.SetFlags().
func (cmd *Update) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.generators, "generators", false, "update generators")
	f.BoolVar(&cmd.prerelease, "prerelease", false, "update to prerelease")
	f.BoolVar(&cmd.force, "force", false, "download latest version even if a new version isn't available")
	f.StringVar(&cmd.ghToken, "ghtoken", "", "Github token for private repos")
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
	} else {
		if code := cmd.updateCLI(); code != subcommands.ExitSuccess {
			return code
		}
	}

	return subcommands.ExitSuccess
}

func (cmd *Update) updateGenerators() subcommands.ExitStatus {
	if cmd.ghToken == "" {
		cmd.ghToken = os.Getenv("GITHUB_TOKEN")
	}

	path, err := generatorsPath()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Find all installed repos.
	matches, err := filepath.Glob(filepath.Join(path, "*", "*", repo.StatesDir, "*"))
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	for _, match := range matches {
		var (
			parts = strings.Split(match, string(filepath.Separator))
			l     = len(parts)
			owner = parts[l-4]
			rep   = parts[l-3]
			ref   = parts[l-1]
			name  = fmt.Sprintf("%s/%s@%s", owner, rep, ref)
			p     = filepath.Join(path, owner, rep)
		)

		fmt.Printf("Updating generators %q...\n", name)

		r := repo.New(p, owner, rep, cmd.ghToken)
		if err != nil {
			fmt.Println(err)
			return subcommands.ExitFailure
		}

		_, updated, err := r.Update(ref, cmd.force)
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

	return subcommands.ExitSuccess
}

func (cmd *Update) updateCLI() subcommands.ExitStatus {
	fmt.Println("Updating CLI...")

	// Get the releases from Github.
	client := github.NewClient(nil)
	rels, res, err := client.Repositories.ListReleases(CLIOwner, CLIRepo, nil)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer res.Body.Close()

	// Find the latest one.
	var (
		asset *github.ReleaseAsset
		tag   *string
	)
	for _, r := range rels {
		if *r.Prerelease == cmd.prerelease {
			if cmd.force || *r.TagName != "v"+cmd.Version {
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
	fmt.Printf("Downloading %q...\n", *asset.Name)

	// Read the archive.
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

	// Create a temporary directory.
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file.
	tempZipFile := filepath.Join(tempDir, "temp.zip")
	f, err := os.OpenFile(tempZipFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer f.Close()

	// Copy the archive to the temporary file.
	if _, err := io.Copy(f, r); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Read the archive.
	zr, err := zip.OpenReader(tempZipFile)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Extracting %q...\n", *asset.Name)

	// This is the name of the binary we want within the archive.
	wantBin := CLIAssetBinary
	if runtime.GOOS == win {
		wantBin = CLIAssetBinaryWin
	}

	// This is the name of the signature we want within the archive.
	wantSig := wantBin + CLISigExt

	// Find the binary and signature in the archive.
	var binZF, sigZF *zip.File
	for _, f := range zr.File {
		switch f.Name {
		case wantBin:
			binZF = f
		case wantSig:
			sigZF = f
		}
	}

	if binZF == nil {
		fmt.Printf("Could not find binary %q in %q.\n", wantBin, *asset.Name)
		return subcommands.ExitFailure
	}
	if sigZF == nil {
		fmt.Printf("Could not find signature %q in %q.\n", wantSig, *asset.Name)
		return subcommands.ExitFailure
	}

	// Get the current binary path.
	execPath, err := osext.Executable()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Read the new binary.
	binRC, err := binZF.Open()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Get the current file info.
	info, err := os.Stat(execPath)
	if err != nil {
		binRC.Close()
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Create and open a file for the new binary.
	newBinPath := filepath.Join(tempDir, filepath.Base(execPath))
	newBin, err := os.OpenFile(newBinPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		binRC.Close()
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Copy the new binary.
	_, err = io.Copy(newBin, binRC)
	binRC.Close()
	newBin.Close()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Read the signature.
	sigRC, err := sigZF.Open()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Create and open a file for the signature.
	sigPath := filepath.Join(tempDir, filepath.Base(execPath)+CLISigExt)
	sig, err := os.OpenFile(sigPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		sigRC.Close()
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Copy the signature.
	_, err = io.Copy(sig, sigRC)
	sigRC.Close()
	sig.Close()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Check the signature.
	fmt.Println("Verifying cryptographic signature...")
	if err := checkSig(newBinPath, sigPath); err != nil {
		fmt.Printf("Failed to verify signature: %s.\n", err)
		return subcommands.ExitFailure
	}

	// Rename old binary.
	// Remove previous old binary if present.
	oldBinPath := filepath.Join(filepath.Dir(execPath), CLIOldBinary)
	if err := os.Remove(oldBinPath); !os.IsNotExist(err) {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if err := os.Rename(execPath, oldBinPath); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Define a function that will try to recover the old binary if anything
	// goes wrong.
	recover := func() {
		if err := os.Remove(execPath); !os.IsNotExist(err) {
			fmt.Println(err)
		}
		if err := os.Rename(oldBinPath, execPath); err != nil {
			fmt.Println(err)
		}
	}

	// Copy new binary to final destination.
	dst, err := os.OpenFile(execPath, os.O_CREATE|os.O_WRONLY, info.Mode())
	if err != nil {
		fmt.Println(err)
		recover()
		return subcommands.ExitFailure
	}

	newBin, err = os.Open(newBinPath)
	if err != nil {
		fmt.Println(err)
		dst.Close()
		recover()
		return subcommands.ExitFailure
	}

	_, err = io.Copy(dst, newBin)
	dst.Close()
	newBin.Close()
	if err != nil {
		fmt.Println(err)
		recover()
		return subcommands.ExitFailure
	}

	fmt.Println("CLI updated successfully.")
	return subcommands.ExitSuccess
}

func checkSig(targetPath, sigPath string) error {
	target, err := os.Open(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()
	sig, err := os.Open(sigPath)
	if err != nil {
		return err
	}
	defer sig.Close()
	r := bytes.NewReader([]byte(pubKey))
	keyring, err := openpgp.ReadArmoredKeyRing(r)
	if err != nil {
		return err
	}
	_, err = openpgp.CheckArmoredDetachedSignature(keyring, target, sig)
	return err
}
