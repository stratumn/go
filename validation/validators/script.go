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
	"context"
	"crypto/sha256"
	"io/ioutil"
	"path"
	"plugin"
	"strings"

	cj "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	golang = "go"

	// ErrLoadingPlugin is the error returned in case the plugin could not be loaded
	ErrLoadingPlugin = "Error while loading validation script for process %s and type %s"

	// ErrBadPlugin is the error returned in case the plugin is missing exported symbols
	ErrBadPlugin = "script does not implement the ScriptValidatorFunc type"

	// ErrBadScriptType is the error returned when the type of script does not match the supported ones
	ErrBadScriptType = "Validation engine does not handle script of type %s, valid types are %v"
)

var (
	// ValidScriptTypes contains the handled languages for validation scripts
	ValidScriptTypes = []string{golang}
)

// ScriptConfig defines the configuration of the go validation plugin.
type ScriptConfig struct {
	File string `json:"file"`
	Type string `json:"type"`
}

// ScriptValidatorFunc is the function called when enforcing a custom validation rule
type ScriptValidatorFunc = func(store.SegmentReader, *cs.Link) error

// ScriptValidator validates a link according to custom rules written as a go plugin.
type ScriptValidator struct {
	script     ScriptValidatorFunc
	ScriptHash types.Bytes32
	Config     *ValidatorBaseConfig
}

func checkScriptType(cfg *ScriptConfig) error {
	switch cfg.Type {
	case golang:
		return nil
	default:
		return errors.Errorf(ErrBadScriptType, cfg.Type, ValidScriptTypes)
	}
}

// NewScriptValidator instanciates a new go plugin and returns a new ScriptValidator.
func NewScriptValidator(baseConfig *ValidatorBaseConfig, scriptCfg *ScriptConfig, pluginsPath string) (Validator, error) {
	if err := checkScriptType(scriptCfg); err != nil {
		return nil, err
	}
	pluginFile := path.Join(pluginsPath, scriptCfg.File)
	p, err := plugin.Open(pluginFile)
	if err != nil {
		return nil, errors.Wrapf(err, ErrLoadingPlugin, baseConfig.Process, baseConfig.LinkType)
	}

	symbol, err := p.Lookup(strings.Title(baseConfig.LinkType))
	if err != nil {
		return nil, errors.Wrapf(err, ErrLoadingPlugin, baseConfig.Process, baseConfig.LinkType)
	}

	customValidator, ok := symbol.(ScriptValidatorFunc)
	if !ok {
		return nil, errors.Wrapf(errors.New(ErrBadPlugin), ErrLoadingPlugin, baseConfig.Process, baseConfig.LinkType)
	}

	// here we ignore the error since there is no way we cannot read the file if the plugin has been loaded successfully
	b, _ := ioutil.ReadFile(pluginFile)
	return &ScriptValidator{
		Config:     baseConfig,
		script:     customValidator,
		ScriptHash: sha256.Sum256(b),
	}, nil
}

// Hash implements github.com/stratumn/go-indigocore/validation/validators.Validator.Hash.
func (sv ScriptValidator) Hash() (*types.Bytes32, error) {
	b, err := cj.Marshal(sv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	validationsHash := types.Bytes32(sha256.Sum256(b))
	return &validationsHash, nil
}

// ShouldValidate implements github.com/stratumn/go-indigocore/validation/validators.Validator.ShouldValidate.
func (sv ScriptValidator) ShouldValidate(link *cs.Link) bool {
	return sv.Config.ShouldValidate(link)
}

// Validate implements github.com/stratumn/go-indigocore/validation/validators.Validator.Validate.
func (sv ScriptValidator) Validate(_ context.Context, storeReader store.SegmentReader, link *cs.Link) error {
	return sv.script(storeReader, link)
}
