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
	fmt.Println("Updating generators...")

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

		fmt.Printf("  * Updating %q...\n", name)

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
			fmt.Printf("  * %q updated successfully.\n", name)
		} else {
			fmt.Printf("  * %q already up-to-date.\n", name)
		}
	}

	fmt.Println("Generators updated successfully.")

	return subcommands.ExitSuccess
}

func (cmd *Update) updateCLI() subcommands.ExitStatus {
	fmt.Println("Updating CLI...")
	client := github.NewClient(nil)

	// Find latest release.
	asset, tag, err := cmd.findRelease(client)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if asset == nil {
		fmt.Println("CLI already up-to-date.")
		return subcommands.ExitSuccess
	}

	fmt.Printf("  * Downloading %q@%q...\n", *asset.Name, *tag)

	// Create temporary directory.
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer os.RemoveAll(tempDir)

	// Download release.
	tempZipFile := filepath.Join(tempDir, "temp.zip")
	if err := dlRelease(client, tempZipFile, asset); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	fmt.Printf("  * Extracting %q...\n", *asset.Name)

	// Find binary and signature.
	zrc, binZF, sigZF, err := findReleaseFiles(tempZipFile)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	defer zrc.Close()
	if binZF == nil {
		fmt.Printf("Could not find binary in %q.\n", *asset.Name)
		return subcommands.ExitFailure
	}
	if sigZF == nil {
		fmt.Printf("Could not find signature in %q.\n", *asset.Name)
		return subcommands.ExitFailure
	}

	// Get the current binary path.
	execPath, err := osext.Executable()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Get the current binary file info.
	info, err := os.Stat(execPath)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Copy the new binary to the temporary directory.
	binPath := filepath.Join(tempDir, filepath.Base(execPath))
	if err := copyZF(binPath, binZF, 0644); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Copy the signature the the temporary directory.
	sigPath := filepath.Join(tempDir, filepath.Base(execPath)+CLISigExt)
	if err := copyZF(sigPath, sigZF, 0644); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Check the signature.
	fmt.Println("  * Verifying cryptographic signature...")
	if err := checkSig(binPath, sigPath); err != nil {
		fmt.Printf("Failed to verify signature: %s.\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println("  * Updating binary...")

	// Remove previous old binary if present.
	oldPath := filepath.Join(filepath.Dir(execPath), CLIOldBinary)
	if err := os.Remove(oldPath); !os.IsNotExist(err) {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Rename current binary.
	if err := os.Rename(execPath, oldPath); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	// Copy new binary to final destination.
	if err := copyF(execPath, binPath, info.Mode()); err != nil {
		fmt.Println(err)
		// Try to recover old binary.
		if err := os.Remove(execPath); !os.IsNotExist(err) {
			fmt.Println(err)
		}
		if err := os.Rename(oldPath, execPath); err != nil {
			fmt.Println(err)
		}
		return subcommands.ExitFailure
	}

	fmt.Println("CLI updated successfully.")
	return subcommands.ExitSuccess
}

func (cmd *Update) findRelease(client *github.Client) (*github.ReleaseAsset, *string, error) {
	rels, res, err := client.Repositories.ListReleases(CLIOwner, CLIRepo, nil)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	var (
		name  = fmt.Sprintf(CLIAssetFormat, runtime.GOOS, runtime.GOARCH)
		asset *github.ReleaseAsset
		tag   *string
	)
	for _, r := range rels {
		if *r.Prerelease == cmd.prerelease {
			if cmd.force || *r.TagName != "v"+cmd.Version {
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

	return asset, tag, nil
}

func dlRelease(client *github.Client, dst string, asset *github.ReleaseAsset) error {
	rc, url, err := client.Repositories.DownloadReleaseAsset(CLIOwner, CLIRepo, *asset.ID)
	if err != nil {
		return err
	}

	var r io.ReadCloser

	if rc != nil {
		r = rc
	} else if url != "" {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		r = res.Body
	}
	defer r.Close()

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func findReleaseFiles(src string) (*zip.ReadCloser, *zip.File, *zip.File, error) {
	rc, err := zip.OpenReader(src)
	if err != nil {
		return nil, nil, nil, err
	}

	wantBin := CLIAssetBinary
	if runtime.GOOS == win {
		wantBin = CLIAssetBinaryWin
	}
	wantSig := wantBin + CLISigExt

	var binZF, sigZF *zip.File
	for _, f := range rc.File {
		switch f.Name {
		case wantBin:
			binZF = f
		case wantSig:
			sigZF = f
		}
	}

	return rc, binZF, sigZF, nil
}

func copy(dst string, r io.Reader, mode os.FileMode) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func copyZF(dst string, zf *zip.File, mode os.FileMode) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	return copy(dst, rc, mode)
}

func copyF(dst, src string, mode os.FileMode) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return copy(dst, f, mode)
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
