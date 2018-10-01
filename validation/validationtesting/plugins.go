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

package validationtesting

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Sample validation plugins.
const (
	// Plugin missing the required Validate func.
	PluginMissingValidate = []byte(`
		package main

		import (
			"context"
		
			"github.com/stratumn/go-chainscript"
			"github.com/stratumn/go-core/store"
		)

		func NotValidate(_ context.Context, _ store.SegmentReader, _ *chainscript.Link) error {
			return nil
		}

		func main() {}
	`)

	// Plugin containing an invalid Validate func (signature doesn't match).
	PluginInvalidValidate = []byte(`
		package main

		import (
			"context"
		
			"github.com/stratumn/go-chainscript"
			"github.com/stratumn/go-core/store"
		)

		func Validate(_ store.SegmentReader, _ *chainscript.Link) error {
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
		
			"github.com/stratumn/go-chainscript"
			"github.com/stratumn/go-core/store"
		)

		func Validate(_ context.Context, _ store.SegmentReader, _ *chainscript.Link) error {
			return errors.New("invalid link")
		}

		func main() {}
	`)

	// Plugin that returns a validation success.
	PluginValidationSuccess = []byte(`
		package main

		import (
			"context"
		
			"github.com/stratumn/go-chainscript"
			"github.com/stratumn/go-core/store"
		)

		func Validate(_ context.Context, _ store.SegmentReader, _ *chainscript.Link) error {
			return nil
		}

		func main() {}
	`)
)

// CompilePlugin compiles the given plugin and stores it in a temporary folder.
// The name of the file will be its hash.
// It returns the path to the compiled file.
func CompilePlugin(t *testing.T, pluginContent []byte) string {
	tmpDir, err := ioutil.TempDir("", "plugins-test")
	require.NoError(t, err)

	sourceFile, err := ioutil.TempFile(tmpDir, "")
	require.NoError(t, err)

	_, err = sourceFile.Write(pluginContent)
	require.NoError(t, err)

	err = sourceFile.Close()
	require.NoError(t, err)

	err = os.Rename(sourceFile, sourceFile+".go")
	require.NoError(t, err)

	pluginFile := strings.Replace(sourceFile, ".go", ".so", 1)
	err := exec.Command("go", "build", "-o", sourceFile, "-buildmode=plugin", pluginFile).Run()

	b, err := ioutil.ReadFile(pluginFile)
	require.NoError(t, err)

	pluginHash := sha256.Sum256(b)
	pluginFinalPath := filepath.Join(tmpDir, fmt.Sprintf("%s.so", pluginHash))
	err = os.Rename(pluginFile, pluginFinalPath)
	require.NoError(t, err)

	return pluginFinalPath
}
