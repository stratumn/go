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

package tmpop

import "encoding/json"

// Query is the type used to query the tendermint App
type Query struct {
	Name string `json:"Name"`
	Args []byte `json:"Args"`
}

// BuildQueryBinary outputs the marshalled Query
func BuildQueryBinary(name string, args interface{}) ([]byte, error) {
	var argsBytes []byte
	if args != nil {
		var err error
		if argsBytes, err = json.Marshal(args); err != nil {
			return nil, err
		}
	}

	query := &Query{Name: name, Args: argsBytes}
	bytes, err := json.Marshal(query)

	if err != nil {
		return nil, err
	}
	return bytes, nil
}
