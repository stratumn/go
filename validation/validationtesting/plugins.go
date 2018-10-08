// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

package validationtesting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Sample validation plugins.
var (
	// Plugin missing the required Validate func.
	PluginMissingValidate = []byte(`
		package main

		import (
			"context"
		
			"github.com/stratumn/go-core/store"
			"github.com/stratumn/go-core/types"
		)

		func NotValidate(_ context.Context, _ store.SegmentReader, _ *types.Link) error {
			return nil
		}

		func main() {}
	`)

	// Plugin containing an invalid Validate func (signature doesn't match).
	PluginInvalidValidate = []byte(`
		package main

		import (
			"github.com/stratumn/go-core/store"
			"github.com/stratumn/go-core/types"
		)

		func Validate(_ store.SegmentReader, _ *types.Link) error {
			return nil
		}

		func main() {}
	`)

	// Plugin that returns a validation failure.
	PluginValidationError = []byte(`
		package main

		import (
			"context"
			"errors"
		
			"github.com/stratumn/go-core/store"
			"github.com/stratumn/go-core/types"
		)

		func Validate(_ context.Context, _ store.SegmentReader, _ *types.Link) error {
			return errors.New("invalid link")
		}

		func main() {}
	`)

	// Plugin that returns a validation success.
	PluginValidationSuccess = []byte(`
		package main

		import (
			"context"
			"errors"
		
			"github.com/stratumn/go-core/store"
			"github.com/stratumn/go-core/types"
		)

		func Validate(_ context.Context, _ store.SegmentReader, l *types.Link) error {
			if l.Link.Meta.Process == nil {
				return errors.New("link is missing process")
			}

			return nil
		}

		func main() {}
	`)
)

// CompilePlugin compiles the given plugin and stores it in a temporary folder.
// The name of the file will be its hex-encoded hash.
// It returns the path to the directory and the compiled file hash.
func CompilePlugin(t *testing.T, pluginContent []byte) (string, []byte) {
	// To get reproducible builds we need the path to the plugin to be stable.
	// There is some effort ongoing to get reproducible builds in Go regardless
	// of the path but it's not yet working.
	pluginsDir := path.Join(os.TempDir(), "go-core-test-plugins")
	err := os.MkdirAll(pluginsDir, os.ModePerm)
	require.NoError(t, err)

	sourceHash := sha256.Sum256(pluginContent)
	sourceFile := path.Join(pluginsDir, fmt.Sprintf("%s.go", hex.EncodeToString(sourceHash[:])))
	err = ioutil.WriteFile(sourceFile, pluginContent, os.ModePerm)
	require.NoError(t, err)

	pluginFile := strings.Replace(sourceFile, ".go", ".so", 1)
	buildCmd := exec.Command("go", "build", "-o", pluginFile, "-buildmode=plugin", sourceFile)
	err = buildCmd.Run()
	require.NoError(t, err)

	b, err := ioutil.ReadFile(pluginFile)
	require.NoError(t, err)

	pluginHash := sha256.Sum256(b)
	pluginFinalName := fmt.Sprintf("%s.so", hex.EncodeToString(pluginHash[:]))
	err = os.Rename(pluginFile, filepath.Join(pluginsDir, pluginFinalName))
	require.NoError(t, err)

	return pluginsDir, pluginHash[:]
}
