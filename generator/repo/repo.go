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

// Package repo deals with a Github repository of generators.
package repo

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-github/github"
	"github.com/stratumn/go/generator"
)

const (
	// TagsDir is the name of the directory where tags are stored.
	TagsDir = "tags"

	// DescPerm if the file mode the a repo description.
	DescPerm = 0644

	// TagPerm is the file mode for a tag directory.
	TagPerm = 0755
)

// Desc descibes a repository.
type Desc struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Tag   string `json:"tag"`
}

// Repo manages a Github repository.
type Repo struct {
	path   string
	owner  string
	repo   string
	client *github.Client
}

// New instantiates a repository.
func New(path, owner, repo string) *Repo {
	return &Repo{
		path:   path,
		owner:  owner,
		repo:   repo,
		client: github.NewClient(nil),
	}
}

// Update download the latest release if needed.
func (r *Repo) Update() (*Desc, bool, error) {
	desc, err := r.GetDesc()
	if err != nil {
		return nil, false, err
	}

	rel, res, err := r.client.Repositories.GetLatestRelease(r.owner, r.repo)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	if desc != nil && desc.Tag == *rel.TagName {
		return desc, false, nil
	}

	desc, err = r.Download(rel)
	if err != nil {
		return nil, false, err
	}

	name := filepath.Join(r.path, "repo.json")
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, DescPerm)
	if err != nil {
		return nil, false, err
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(desc); err != nil {
		return nil, false, err
	}

	return desc, true, nil
}

// GetDesc returns the description of the repository.
// If the repository does not exist, it returns nil.
func (r *Repo) GetDesc() (*Desc, error) {
	name := filepath.Join(r.path, "repo.json")
	var desc *Desc
	f, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	desc = &Desc{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&desc); err != nil {
		return nil, err
	}
	return desc, err
}

// GetDescOrCreate returns the description of the repository.
// If the repository does not exist, it returns creates it by calling Update().
func (r *Repo) GetDescOrCreate() (*Desc, error) {
	desc, err := r.GetDesc()
	if err != nil {
		return nil, err
	}
	if desc == nil {
		if desc, _, err = r.Update(); err != nil {
			return nil, err
		}
	}
	return desc, nil
}

// List lists the generators of the repository.
func (r *Repo) List() ([]*generator.Definition, error) {
	desc, err := r.GetDescOrCreate()
	if err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(filepath.Join(r.path, TagsDir, desc.Tag, "*", "generator.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)

	var defs []*generator.Definition
	for _, p := range matches {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		var def generator.Definition
		if err = dec.Decode(&def); err != nil {
			return nil, err
		}
		defs = append(defs, &def)
	}

	return defs, nil
}

// Generate executes a generator by name.
func (r *Repo) Generate(name, dst string, opts *generator.Options) error {
	desc, err := r.GetDescOrCreate()

	matches, err := filepath.Glob(filepath.Join(r.path, TagsDir, desc.Tag, "*", "generator.json"))
	if err != nil {
		return err
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

		if def.Name == name {
			gen, err := generator.NewFromDir(filepath.Dir(p), opts)
			if err != nil {
				return err
			}
			return gen.Exec(dst)
		}
	}

	return fmt.Errorf("could not find generator %q", name)
}

// Download downloads a release.
func (r *Repo) Download(release *github.RepositoryRelease) (*Desc, error) {
	res, err := http.Get(*release.TarballURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
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
			parts := strings.Split(hdr.Name, string(filepath.Separator))
			parts[0] = *release.TagName
			dst := filepath.Join(r.path, TagsDir, filepath.Join(parts...))
			err = os.MkdirAll(filepath.Dir(dst), TagPerm)
			if err != nil {
				return nil, err
			}
			mode := hdr.FileInfo()
			f, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Mode())
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(f, tr)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Desc{
		Owner: r.owner,
		Repo:  r.repo,
		Tag:   *release.TagName,
	}, nil
}
