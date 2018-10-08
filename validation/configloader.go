// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/stratumn/go-core/validation/validators"
)

// LoadFromFile loads the validation rules from a json file.
func LoadFromFile(validationCfg *Config) (validators.ProcessesValidators, error) {
	f, err := os.Open(validationCfg.RulesPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var rules ProcessesRules
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return rules.Validators(validationCfg.PluginsPath)
}
