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

// TODO: deal with context properly.

package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/google/go-github/github"
	"github.com/kardianos/osext"
	"github.com/spf13/cobra"
	"github.com/stratumn/sdk/generator/repo"
)

var (
	force      bool
	prerelease bool
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Stratumn CLI and generators",
	Long: `Update Stratumn CLI and update generators to latest version.

It downloads the latest version of the Stratumn CLI if available. It checks that the binary is cryptographically signed before installing.

It will also update the generators if newer versions are available.	
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("unexpected arguments")
		}

		if isBrew() && !force {
			fmt.Println("NOTE: The CLI won't be updated because a Homebrew installation was detected. To update the CLI to the latest version, run `brew update && brew upgrade`.")
		} else if err := updateCLI(); err != nil {
			return err
		}

		return updateGenerators()
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.PersistentFlags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Download latest version even if not more recent",
	)

	updateCmd.PersistentFlags().BoolVarP(
		&prerelease,
		"prerelease",
		"P",
		false,
		"Download prerelease version",
	)
}

func isBrew() bool {
	if runtime.GOOS != darwin {
		return false
	}

	fmt.Println("Checking for Homebrew installation...")

	out, err := exec.Command(brewInfoCmd[0], brewInfoCmd[1:]...).Output()
	if err != nil {
		return false
	}

	var info []struct {
		LinkedKeg string `json:"linked_keg"`
	}

	if err := json.Unmarshal(out, &info); err != nil {
		return false
	}

	for _, i := range info {
		if i.LinkedKeg == "v"+version {
			return true
		}
	}

	return false
}

func updateGenerators() error {
	fmt.Println("Updating generators...")

	// Find all installed repos.
	path := generatorsPath
	matches, err := filepath.Glob(filepath.Join(path, "*", "*", repo.StatesDir, "*"))
	if err != nil {
		return err
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

		r := repo.New(p, owner, rep, ghToken, !generatorsUseLocalFiles)
		if err != nil {
			return err
		}

		_, updated, err := r.Update(ref, force)
		if err != nil {
			return err
		}

		if updated {
			fmt.Printf("  * %q updated successfully.\n", name)
		} else {
			fmt.Printf("  * %q already up-to-date.\n", name)
		}
	}

	fmt.Println("Generators updated successfully.")

	return nil
}

func updateCLI() error {
	fmt.Println("Updating CLI...")
	client := github.NewClient(nil)

	// Find latest release.
	asset, tag, err := findRelease(client)
	if err != nil {
		return err
	}
	if asset == nil {
		fmt.Println("CLI already up-to-date.")
		return nil
	}

	fmt.Printf("  * Downloading %q@%q...\n", *asset.Name, *tag)

	// Create temporary directory.
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Download release.
	tempZipFile := filepath.Join(tempDir, "temp.zip")
	if err = dlRelease(client, tempZipFile, asset); err != nil {
		return err
	}

	fmt.Printf("  * Extracting %q...\n", *asset.Name)

	// Find binary and signature.
	zrc, binZF, sigZF, err := findReleaseFiles(tempZipFile)
	if err != nil {
		return err
	}
	defer zrc.Close()
	if binZF == nil {
		return fmt.Errorf("could not find binary in %q", *asset.Name)
	}
	if sigZF == nil {
		return fmt.Errorf("could not find signature in %q", *asset.Name)
	}

	// Get the current binary path.
	execPath, err := osext.Executable()
	if err != nil {
		return err
	}

	// Get the current binary file info.
	info, err := os.Stat(execPath)
	if err != nil {
		return err
	}

	// Copy the new binary to the temporary directory.
	binPath := filepath.Join(tempDir, filepath.Base(execPath))
	if err := copyZF(binPath, binZF, 0644); err != nil {
		return err
	}

	// Copy the signature the the temporary directory.
	sigPath := filepath.Join(tempDir, filepath.Base(execPath)+SigExt)
	if err := copyZF(sigPath, sigZF, 0644); err != nil {
		return err
	}

	// Check the signature.
	fmt.Println("  * Verifying cryptographic signature...")
	if err := checkSig(binPath, sigPath); err != nil {
		return fmt.Errorf("failed to verify signature: %s", err)
	}

	fmt.Println("  * Updating binary...")

	// Remove previous old binary if present.
	oldPath := filepath.Join(filepath.Dir(execPath), OldBinary)
	if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Rename current binary.
	if err := os.Rename(execPath, oldPath); err != nil {
		return err
	}

	// Copy new binary to final destination.
	if err := copyF(execPath, binPath, info.Mode()); err != nil {
		fmt.Println(err)
		// Try to recover old binary.
		if err := os.Remove(execPath); err != nil && !os.IsNotExist(err) {
			fmt.Println(err)
		}
		if err := os.Rename(oldPath, execPath); err != nil {
			fmt.Println(err)
		}
		return errors.New("failed to update binary")
	}

	fmt.Println("CLI updated successfully.")

	return nil
}

func findRelease(client *github.Client) (*github.ReleaseAsset, *string, error) {
	rels, res, err := client.Repositories.ListReleases(context.TODO(), Owner, Repo, nil)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	var (
		name  = fmt.Sprintf(AssetFormat, runtime.GOOS, runtime.GOARCH)
		asset *github.ReleaseAsset
		tag   *string
	)
	for _, r := range rels {
		if *r.Prerelease == prerelease {
			if force || *r.TagName != "v"+version {
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
	release, _, err := client.Repositories.GetReleaseAsset(context.TODO(), Owner, Repo, *asset.ID)
	if err != nil {
		return err
	}

	res, err2 := http.Get(*release.BrowserDownloadURL)
	if err2 != nil {
		return err2
	}
	r := res.Body
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

	wantBin := AssetBinary
	if runtime.GOOS == win {
		wantBin = AssetBinaryWin
	}
	wantSig := wantBin + SigExt

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
