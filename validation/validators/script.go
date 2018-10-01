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

package validators

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
	"plugin"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// Errors returned by the script validator.
var (
	ErrLoadingPlugin     = errors.New("could not load validation script")
	ErrInvalidPlugin     = errors.New("script does not expose a 'Validate' ScriptValidatorFunc")
	ErrInvalidPluginHash = errors.New("script digest doesn't match received file")
)

// ScriptConfig defines the configuration of the go validation plugin.
type ScriptConfig struct {
	Hash []byte `json:"hash"`
}

// Link simply wraps a link because directly using a chainscript.Link doesn't
// work in plugins (because the plugin compilation uses GOPATH instead of the
// vendor directory).
// Even if your GOPATH and vendor directory have the same version of
// go-chainscript the runtime will treat them as different types so the cast
// to a ScriptValidatorFunc will fail.
type Link struct {
	Link *chainscript.Link
}

// ScriptValidatorFunc is the function called when enforcing a custom
// validation rule.
type ScriptValidatorFunc = func(context.Context, store.SegmentReader, *Link) error

// ScriptValidator validates a link according to custom rules written as a go
// plugin. The plugin should expose a `Validate` method.
type ScriptValidator struct {
	process    string
	script     ScriptValidatorFunc
	scriptHash []byte
}

// NewScriptValidator creates a new validator for the given process.
// It expects a plugin named `{hash}.so` to be found in the pluginsPath
// directory (where {hash} is hex-encoded).
// The plugin should expose a `Validate` ScriptValidatorFunc.
func NewScriptValidator(process string, pluginsPath string, scriptCfg *ScriptConfig) (Validator, error) {
	pluginFile := path.Join(pluginsPath, fmt.Sprintf("%s.so", hex.EncodeToString(scriptCfg.Hash)))
	pluginBytes, err := ioutil.ReadFile(pluginFile)
	if err != nil {
		return nil, errors.Wrap(err, ErrLoadingPlugin.Error())
	}

	fileHash := sha256.Sum256(pluginBytes)
	if !bytes.Equal(scriptCfg.Hash, fileHash[:]) {
		return nil, errors.Wrap(ErrInvalidPlugin, ErrInvalidPluginHash.Error())
	}

	p, err := plugin.Open(pluginFile)
	if err != nil {
		return nil, errors.Wrap(err, ErrLoadingPlugin.Error())
	}

	validateSymbol, err := p.Lookup("Validate")
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidPlugin.Error())
	}

	validate, ok := validateSymbol.(ScriptValidatorFunc)
	if !ok {
		return nil, errors.Wrap(ErrInvalidPlugin, ErrInvalidPlugin.Error())
	}

	return &ScriptValidator{
		process:    process,
		script:     validate,
		scriptHash: fileHash[:],
	}, nil
}

// Hash of the script validator.
func (sv *ScriptValidator) Hash() (*types.Bytes32, error) {
	validatorHash := types.Bytes32(sha256.Sum256(append([]byte(sv.process), sv.scriptHash...)))
	return &validatorHash, nil
}

// ShouldValidate checks that the process matches.
func (sv *ScriptValidator) ShouldValidate(link *chainscript.Link) bool {
	return sv.process == link.Meta.Process.Name
}

// Validate the link.
func (sv *ScriptValidator) Validate(ctx context.Context, storeReader store.SegmentReader, link *chainscript.Link) error {
	return sv.script(ctx, storeReader, &Link{Link: link})
}
