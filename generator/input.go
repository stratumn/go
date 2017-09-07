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

// Input must be implemented by all input types.
import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// StringInputID is the string identifying a string input.
	StringInputID = "string"

	// StringSelectID is the string identifying a string select.
	StringSelectID = "select:string"

	// StringSliceID is a slice of string for mutiple entries.
	StringSliceID = "slice:string"
)

const noValue = "<no value>"

// Input must be implemented by all input types.
type Input interface {
	// Set must set the value of the input or return an error.
	// It should be able to, at the very least, set the value from a string.
	Set(interface{}) error

	// Get must return the value of the input.
	Get() interface{}

	// Msg must return a message that will be displayed when prompting the
	// value.
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
	inputs := InputMap{}
	for k, v := range raw {
		in, err := UnmarshalJSONInput(v)
		if err != nil {
			return err
		}
		inputs[k] = in
	}
	*im = inputs
	return nil
}

// UnmarshalJSONInput creates an input from JSON.
func UnmarshalJSONInput(data []byte) (Input, error) {
	var shared InputShared
	if err := json.Unmarshal(data, &shared); err != nil {
		return nil, err
	}
	switch shared.Type {
	case StringInputID:
		var in StringInput
		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}
		return &in, nil
	case StringSelectID:
		var in StringSelect
		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}
		return &in, nil
	case StringSliceID:
		var in = StringSlice{Separator: ","}
		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}
		return &in, nil
	default:
		return nil, fmt.Errorf("invalid input type %q", shared.Type)
	}
}

// InputShared contains properties shared by all input types.
type InputShared struct {
	// Type is the type of the input.
	Type string `json:"type"`

	// Prompt is the string that will be displayed to the user when asking
	// the value.
	Prompt string `json:"prompt"`
}

// StringInput contains properties for string inputs.
type StringInput struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Format is a string containing a regexp the value must have.
	Format string `json:"format"`

	value string
}

// Set implements github.com/stratumn/sdk/generator.Input.
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

// Get implements github.com/stratumn/sdk/generator.Input.
func (in StringInput) Get() interface{} {
	if in.value == "" && in.Default != noValue {
		return in.Default
	}
	return in.value
}

// Msg implements github.com/stratumn/sdk/generator.Input.
func (in *StringInput) Msg() string {
	if in.Default != "" && in.Default != noValue {
		return fmt.Sprintf("%s (default %q)\n", in.Prompt, in.Default)
	}
	return in.Prompt + "\n"
}

// StringSelect contains properties for string select inputs.
type StringSelect struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Options is an array of possible values.
	Options []StringSelectOption `json:"options"`

	value string
}

// Set implements github.com/stratumn/sdk/generator.Input.
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

// Get implements github.com/stratumn/sdk/generator.Input.
func (in StringSelect) Get() interface{} {
	if in.value == "" && in.Default != noValue {
		return in.Default
	}
	return in.value
}

// Msg implements github.com/stratumn/sdk/generator.Input.
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
	// Input is the string the user must enter to choose this option.
	Input string `json:"input"`

	// Value is the value the input will have if this option is selected.
	Value string `json:"value"`

	// Text will be displayed when presenting this option to the user.
	Text string `json:"text"`
}

// StringSlice contains properties for string inputs.
type StringSlice struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Format is a string containing a regexp the value must have.
	Format string `json:"format"`

	// Separator is a string used to split the input to list.
	Separator string `json:"separator"`

	values []string
}

// Set implements github.com/stratumn/sdk/generator.Input.
func (in *StringSlice) Set(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return errors.New("value must be a string")
	}
	if str == "" && in.Default != noValue {
		str = in.Default
	}
	if str == "" {
		return fmt.Errorf("list must be non empty")
	}

	for _, value := range strings.Split(str, in.Separator) {
		value = strings.TrimSpace(value)
		if in.Format != "" {
			ok, err := regexp.MatchString(in.Format, value)
			if !ok {
				err = fmt.Errorf("value %q must have format %q", value, in.Format)
			}
			if err != nil {
				in.values = nil
				return err
			}
		}
		in.values = append(in.values, value)
	}
	return nil
}

// Get implements github.com/stratumn/sdk/generator.Input.
func (in StringSlice) Get() interface{} {
	if len(in.values) == 0 && in.Default != noValue && in.Separator != "" {
		return strings.Split(in.Default, in.Separator)
	}
	return in.values
}

// Msg implements github.com/stratumn/sdk/generator.Input.
func (in *StringSlice) Msg() string {
	ret := fmt.Sprintf("%s (separator %q)\n", in.Prompt, in.Separator)
	if in.Default != "" && in.Default != noValue {
		ret = fmt.Sprintf("%s (default %q)\n", ret[0:len(ret)-1], in.Default)
	}
	return ret
}
