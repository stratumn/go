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

package generator

// Input must be implemented by all input types.
import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

const (
	// StringInputID is the string identifying a string input.
	StringInputID = "string"

	// StringSelectID is the string identifying a string select.
	StringSelectID = "select:string"
)

const noValue = "<no value>"

// Input must be implemented by all input types.
type Input interface {
	// Set must set the value of the input or return an error.
	Set(interface{}) error

	// Get must return the value of the input.
	Get() interface{}

	// Msg must return a message that will be displayed when prompting the value.
	Msg() string
}

// InputMap is a maps input names to inputs.
type InputMap map[string]Input

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (im *InputMap) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var s map[string]InputShared
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*im = InputMap{}
	for k, v := range s {
		switch v.Type {
		case StringInputID:
			var in StringInput
			if err := json.Unmarshal(raw[k], &in); err != nil {
				return err
			}
			(*im)[k] = &in
		case StringSelectID:
			var in StringSelect
			if err := json.Unmarshal(raw[k], &in); err != nil {
				return err
			}
			(*im)[k] = &in
		default:
			return fmt.Errorf("invalid type %q for input %q", v.Type, k)
		}
	}
	return nil
}

// InputShared contains properties shared by all input types.
type InputShared struct {
	Type   string `json:"type"`
	Prompt string `json:"prompt"`
}

// StringInput contains properties for string inputs.
type StringInput struct {
	InputShared
	Default string `json:"default"`
	Format  string `json:"format"`
	value   string
}

// Set implements github.com/stratumn/go/generator.Input.
func (in *StringInput) Set(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return errors.New("value must be a string")
	}
	if str == "" && in.Default != noValue {
		str = in.Default
	}
	if in.Format != "" {
		ok, err := regexp.MatchString(in.Format, str)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("value must have format %q", in.Format)
		}
	}
	in.value = str
	return nil
}

// Get implements github.com/stratumn/go/generator.Input.
func (in *StringInput) Get() interface{} {
	if in.value == "" && in.Default != noValue {
		return in.Default
	}
	return in.value
}

// Msg implements github.com/stratumn/go/generator.Prompt.
func (in *StringInput) Msg() string {
	if in.Default != "" && in.Default != noValue {
		return in.Prompt + " (default " + in.Default + ")" + "\n"
	}
	return in.Prompt + "\n"
}

// StringSelect contains properties for string select inputs.
type StringSelect struct {
	InputShared
	Default string               `json:"default"`
	Options []StringSelectOption `json:"options"`
	value   string
}

// Set implements github.com/stratumn/go/generator.Input.
func (in *StringSelect) Set(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("value must be a string, got %q", val)
	}
	if str == "" && in.Default != noValue {
		for _, opt := range in.Options {
			if opt.Value == in.Default {
				in.value = opt.Value
				return nil
			}
		}
	}
	for _, opt := range in.Options {
		if opt.Input == str {
			in.value = opt.Value
			return nil
		}
	}
	return fmt.Errorf("invalid value %q", str)
}

// Get implements github.com/stratumn/go/generator.Input.
func (in *StringSelect) Get() interface{} {
	if in.value == "" && in.Default != noValue {
		return in.Default
	}
	return in.value
}

// Msg implements github.com/stratumn/go/generator.Prompt.
func (in *StringSelect) Msg() string {
	p := in.Prompt + "\n"
	for _, v := range in.Options {
		if in.Default == v.Value && in.Default != noValue {
			p += v.Input + ": " + v.Text + " (default)\n"
		} else {
			p += v.Input + ": " + v.Text + "\n"
		}
	}
	return p
}

// StringSelectOption contains properties for string select options.
type StringSelectOption struct {
	Input string `json:"input"`
	Value string `json:"value"`
	Text  string `json:"text"`
}
