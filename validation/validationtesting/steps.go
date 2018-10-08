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
	"encoding/json"

	"github.com/stratumn/go-core/validation"
)

// ChatJSONStepsRules is a set of steps validation rules for the chat process.
const ChatJSONStepsRules = `
{
	"message": {
		"signatures": null,
		"schema": {
			"type": "object",
			"properties": {
				"to": {
					"type": "string"
				},
				"content": {
					"type": "string"
				}
			},
			"required": ["to", "content"]
		},
		"transitions": ["init", "message"]
	},
	"init": {
		"signatures": ["manager", "it"],
		"transitions": [""]
	}
}
`

// AuctionJSONStepsRules is a set of steps validation rules for the auction
// process.
const AuctionJSONStepsRules = `
{
	"init": {
		"signatures": ["alice.vandenbudenmayer@stratumn.com"],
		"schema": {
			"type": "object",
			"properties": {
				"seller": {
					"type": "string"
				},
				"lot": {
					"type": "string"
				},
				"initialPrice": {
					"type": "integer",
					"minimum": 0
				}
			},
			"required": ["seller", "lot", "initialPrice"]
		},
		"transitions": [""]
	},
	"bid": {
		"schema": {
			"type": "object",
			"properties": {
				"buyer": {
					"type": "string"
				},
				"bidPrice": {
					"type": "integer",
					"minimum": 0
				}
			},
			"required": ["buyer", "bidPrice"]
		},
		"transitions": ["init", "bid"]
	}
}
`

// UnmarshalStepsRules unmarshals JSON-formatted steps validation rules.
func UnmarshalStepsRules(data []byte) (map[string]*validation.StepRules, error) {
	var stepsRules map[string]*validation.StepRules
	if err := json.Unmarshal(data, &stepsRules); err != nil {
		return nil, err
	}

	return stepsRules, nil
}
