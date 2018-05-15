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

package validation

import (
	"context"

	"github.com/stratumn/go-indigocore/validation/validators"
)

const (
	// DefaultFilename is the default filename for the file with the rules of validation
	DefaultFilename = "/data/validation/rules.json"

	// DefaultPluginsDirectory is the default directory where validation plugins are located
	DefaultPluginsDirectory = "/data/validation/"
)

// Config contains the path of the rules JSON file and the directory where the validator scripts are located.
type Config struct {
	RulesPath   string
	PluginsPath string
}

// Manager defines the methods to implement to manage validations in an indigo network.
type Manager interface {
	UpdateSubscriber

	// ListenAndUpdate will update the current validators whenever a change occurs in the governance rules.
	// This method must be run in a goroutine as it will wait for events from the network or file updates.
	ListenAndUpdate(ctx context.Context) error

	// Current returns the current version of the validator set.
	Current() validators.Validator
}
