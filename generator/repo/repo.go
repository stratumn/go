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

// Package repo deals with a Github repository of generators.
//
// It provides functionality to store and update remote generators from
// a Github repository locally. It can track a Git branch, a tag, or a
// commit SHA1.
//
// It uses the Github API and doesn't rely on Git.
package repo

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/stratumn/sdk/generator"
)

const (
	// StatesDir is the name of the states directory.
	StatesDir = "states"

	// StateFile is the name of the state file.
	StateFile = "repo.json"

	// StateDirPerm is the file mode for a state directory.
	StateDirPerm = 0755

	// StateFilePerm is the file mode for a state file.
	StateFilePerm = 0644

	// SrcDir is the name of the directory where sources are stored.
	SrcDir = "src"

	// SrcPerm is the file mode for a state directory.
	SrcPerm = 0755
)

// State describes a repository.
type State struct {
	// Owner is the Github username of the owner of the repository.
	Owner string `json:"owner"`

	// Repo is the name of the Github repository.
	Repo string `json:"repo"`

	// Ref is a branch, a tag, or a commit SHA1.
	Ref string `json:"ref"`

	// SHA1 is the commit SHA1 of the downloaded version.
	// It is used to check if an update is available on Github.
	SHA1 string `json:"sha1"`
}

// Repo manages a Github repository.
type Repo struct {
	path     string
	owner    string
	repo     string
	isRemote bool
	client   *github.Client
}

// New instantiates a repository.
func New(path, owner, repo, ghToken string, isRemote bool) *Repo {
	var tc *http.Client
	if ghToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ghToken},
		)
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}
	return &Repo{
		path:     path,
		owner:    owner,
		repo:     repo,
		isRemote: isRemote,
		client:   github.NewClient(tc),
	}
}

// Update downloads the latest release if needed (or if force is true).
// Ref can be a branch, a tag, or a commit SHA1.
func (r *Repo) Update(ref string, force bool) (*State, bool, error) {
	if !r.isRemote {
		return nil, false, nil
	}

	state, err := r.GetState(ref)
	if err != nil {
		return nil, false, err
	}

	sha1 := ""
	if !force && state != nil {
		sha1 = state.SHA1
	}

	sha1, res, err := r.client.Repositories.GetCommitSHA1(context.TODO(), r.owner, r.repo, ref, sha1)
	if res != nil {
		defer res.Body.Close()
		if res.StatusCode == http.StatusNotModified {
			// No update is available.
			return state, false, nil
		}
	}
	if err != nil {
		return nil, false, err
	}

	state, err = r.download(ref, sha1)
	if err != nil {
		return nil, false, err
	}

	path := filepath.Join(r.path, StatesDir, ref, StateFile)
	if err := os.MkdirAll(filepath.Dir(path), StateDirPerm); err != nil {
		return nil, false, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, StateFilePerm)
	if err != nil {
		return nil, false, err
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(state); err != nil {
		return nil, false, err
	}

	return state, true, nil
}

// GetState returns the state of the repository.
// Ref can be a branch, a tag, or a commit SHA1.
// If the repository does not exist locally, it returns nil.
func (r *Repo) GetState(ref string) (*State, error) {
	if !r.isRemote {
		return nil, nil
	}
	path := filepath.Join(r.path, StatesDir, ref, StateFile)
	var state *State
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	state = &State{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}
	return state, err
}

// GetStateOrCreate returns the state of the repository.
// If the repository does not exist locally, it creates it by calling Update().
// Ref can be a branch, a tag, or a commit SHA1.
func (r *Repo) GetStateOrCreate(ref string) (*State, error) {
	state, err := r.GetState(ref)
	if err != nil {
		return nil, err
	}
	if state == nil {
		if state, _, err = r.Update(ref, false); err != nil {
			return nil, err
		}
	}
	return state, nil
}

// createGeneratorPath returns the path to generators files
func (r Repo) createGeneratorPath(ref string) string {
	if r.isRemote {
		return filepath.Join(r.path, SrcDir, ref)
	}
	return r.path
}

// applyFuncOnGenerators iterates on generators and apply a function on all of theses
func (r Repo) applyFuncOnGenerators(ref string, functor func(*generator.Definition, string)) error {
	_, err := r.GetStateOrCreate(ref)
	if err != nil {
		return err
	}

	pattern := filepath.Join(r.createGeneratorPath(ref), "*", generator.DefinitionFile)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return errors.New("No generator found in " + pattern)
	}
	sort.Strings(matches)

	for _, p := range matches {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		var def generator.Definition
		if err = dec.Decode(&def); err != nil {
			return err
		}
		functor(&def, p)
	}

	return nil
}

// List lists the generators of the repository.
// If the repository does not exist locally, it creates it by calling Update().
// Ref can be a branch, a tag, or a commit SHA1.
func (r *Repo) List(ref string) ([]*generator.Definition, error) {
	var defs []*generator.Definition
	err := r.applyFuncOnGenerators(ref, func(def *generator.Definition, _ string) {
		defs = append(defs, def)
	})
	return defs, err
}

// Generate executes a generator by name.
// Ref can be a branch, a tag, or a commit SHA1.
func (r *Repo) Generate(name, dst string, opts *generator.Options, ref string) error {
	var genFound = false
	var genError error
	if err := r.applyFuncOnGenerators(ref, func(def *generator.Definition, path string) {
		if def.Name == name {
			genFound = true
			gen, err := generator.NewFromDir(filepath.Dir(path), opts)
			if err != nil {
				genError = err
			} else {
				genError = gen.Exec(dst)
			}
		}
	}); err != nil {
		return err
	}

	if genFound {
		return genError
	}
	return fmt.Errorf("could not find generator %q", name)
}

func (r *Repo) download(ref, sha1 string) (*State, error) {
	opts := github.RepositoryContentGetOptions{Ref: sha1}
	url, ghres, err := r.client.Repositories.GetArchiveLink(
		context.TODO(),
		r.owner,
		r.repo,
		github.Tarball,
		&opts,
	)
	if err != nil {
		return nil, err
	}
	defer ghres.Body.Close()

	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, err
	}

	if err := os.RemoveAll(filepath.Join(r.path, SrcDir, ref)); err != nil {
		return nil, err
	}

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hdr.Typeflag == tar.TypeReg {
			parts := strings.Split(hdr.Name, "/")
			parts = parts[1:]
			dst := filepath.Join(r.path, SrcDir, ref, filepath.Join(parts...))
			err = os.MkdirAll(filepath.Dir(dst), SrcPerm)
			if err != nil {
				return nil, err
			}
			mode := hdr.FileInfo()
			f, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Mode())
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(f, tr)
			f.Close()
			if err != nil {
				return nil, err
			}
		}
	}

	return &State{
		Owner: r.owner,
		Repo:  r.repo,
		Ref:   ref,
		SHA1:  sha1,
	}, nil
}
