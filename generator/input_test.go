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

package generator

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputSliceUnmarshalJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		f, err := os.Open("testdata/nodejs/generator.json")
		assert.NoError(t, err, "os.Open()")

		dec := json.NewDecoder(f)
		var gen Definition
		err = dec.Decode(&gen)
		assert.NoError(t, err, "dec.Decode()")
		assert.Equal(t, gen.Name, "nodejs", "gen.Name")
		assert.IsType(t, &StringInput{}, gen.Inputs["name"], `gen.Inputs["name"].Msg()`)
		// assert.IsType(t, StringSlice{}, gen.Inputs["name"],  "Project name: (default \"{{.dir}}\")\n", `gen.Inputs["name"].Msg()`)
	})

	t.Run("Invalid", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": [1, 2, 3]}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid input", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": 1}}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid type", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "nope"}}}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid string", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "string", "default": 1}}}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid int", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "int", "default": "1"}}}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid select", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "select:string", "options": [1]}}}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid select multi", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "selectmulti:string", "options": [1]}}}`), &gen)
		assert.Error(t, err, "invalid json")
	})

	t.Run("Invalid slice", func(t *testing.T) {
		var gen Definition
		err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "slice:string", "format": 42}}}`), &gen)
		assert.Error(t, err, "invalid json")
	})
}

func TestStringSelect_CreateItems(t *testing.T) {
	type fields struct {
		InputShared InputShared
		Default     string
		Options     StringSelectOptions
	}
	tests := []struct {
		name      string
		fields    fields
		wantItems []interface{}
	}{
		{
			name: "empty",
			fields: fields{
				Default: "",
				Options: map[string]string{},
			},
			wantItems: []interface{}{},
		},
		{
			name: "sorted",
			fields: fields{
				Default: "",
				Options: map[string]string{"a": "A", "b": "B", "c": "C"},
			},
			wantItems: []interface{}{"A", "B", "C"},
		},
		{
			name: "reverse",
			fields: fields{
				Default: "",
				Options: map[string]string{"c": "C", "b": "B", "a": "A"},
			},
			wantItems: []interface{}{"A", "B", "C"},
		},
		{
			name: "case sensitive",
			fields: fields{
				Default: "",
				Options: map[string]string{"a": "a", "b": "B", "c": "C"},
			},
			wantItems: []interface{}{"B", "C", "a"},
		},
		{
			name: "reverse with first default",
			fields: fields{
				Default: "a",
				Options: map[string]string{"c": "C", "b": "B", "a": "A"},
			},
			wantItems: []interface{}{"A", "B", "C"},
		},
		{
			name: "reverse with last default",
			fields: fields{
				Default: "c",
				Options: map[string]string{"c": "C", "b": "B", "a": "A"},
			},
			wantItems: []interface{}{"C", "A", "B"},
		},
		{
			name: "reverse with unknown default",
			fields: fields{
				Default: "foo",
				Options: map[string]string{"c": "C", "b": "B", "a": "A"},
			},
			wantItems: []interface{}{"A", "B", "C"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &StringSelect{
				InputShared: tt.fields.InputShared,
				Default:     tt.fields.Default,
				Options:     tt.fields.Options,
			}
			assert.EqualValues(t, tt.wantItems, in.CreateItems())
		})
	}
}
