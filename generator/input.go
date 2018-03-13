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
	"fmt"
	"regexp"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

const (
	// IntInputID is the string identifying an int input.
	IntInputID = "int"

	// StringInputID is the string identifying a string input.
	StringInputID = "string"

	// StringSelectID is the string identifying a string select.
	StringSelectID = "select:string"

	// StringSelectMultiID is the string identifying a string select with multiple choices.
	StringSelectMultiID = "selectmulti:string"

	// StringSliceID is a slice of string for mutiple entries.
	StringSliceID = "slice:string"
)

const noValue = "<no value>"

// Input must be implemented by all input types.
type Input interface {
	Run() (interface{}, error)
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
			return errors.Wrapf(err, "cannot unmarshall input file %s", v)
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
	case IntInputID:
		var in IntInput
		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}
		return &in, nil
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
	case StringSelectMultiID:
		var in StringSelectMulti
		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}
		return &in, nil
	case StringSliceID:
		var in = StringSlice{}
		if err := json.Unmarshal(data, &in); err != nil {
			return nil, err
		}
		return &in, nil
	default:
		return nil, errors.Errorf("invalid input type %q", shared.Type)
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

// IntInput contains properties for int inputs.
type IntInput struct {
	InputShared

	Default int `json:"default"`
}

func (in *InputShared) createStringPrompt(label, format, defaultValue string) promptui.Prompt {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if format != "" {
				ok, err := regexp.MatchString(format, input)
				if err != nil {
					return err
				}
				if !ok {
					return errors.Errorf("value must have format %q", format)
				}
			}
			return nil
		},
	}
	if defaultValue != noValue {
		prompt.Default = defaultValue
	}
	return prompt
}

// Run implements github.com/stratumn/sdk/generator.Input.
func (in *IntInput) Run() (interface{}, error) {
	prompt := in.createStringPrompt(in.Prompt, "^[0-9]+$", fmt.Sprintf("%d", in.Default))
	txt, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	i, err := strconv.ParseInt(txt, 10, 0)
	return int(i), err
}

// StringInput contains properties for string inputs.
type StringInput struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Format is a string containing a regexp the value must have.
	Format string `json:"format"`
}

// Run implements github.com/stratumn/sdk/generator.Input.
func (in *StringInput) Run() (interface{}, error) {
	prompt := in.createStringPrompt(in.Prompt, in.Format, in.Default)
	return prompt.Run()
}

// StringSelect contains properties for string select inputs.
type StringSelect struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Options is an array of possible values.
	Options StringSelectOptionSlice `json:"options"`
}

// Run implements github.com/stratumn/sdk/generator.Input.
func (in *StringSelect) Run() (interface{}, error) {
	prompt := promptui.Select{
		Label: in.Prompt,
		Items: func() (items []interface{}) {
			items = make([]interface{}, 0, len(in.Options))
			if in.Default != "" {
				items = append(items, in.Options.FindText(in.Default))
			}
			for _, v := range in.Options {
				if in.Default == "" || v.Value != in.Default {
					items = append(items, v.Text)
				}
			}
			return
		}(),
		Size: len(in.Options),
	}
	_, txt, err := prompt.Run()
	return in.Options.FindValue(txt), err
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

// StringSelectOptionSlice is a slice of StringSelectOption to add methods
type StringSelectOptionSlice []StringSelectOption

// FindText have to be replaced when []StringSelectOption will be a map[string]string
func (options StringSelectOptionSlice) FindText(value string) string {
	for _, v := range options {
		if value == v.Value {
			return v.Text
		}
	}
	return value
}

// FindValue have to be replaced when []StringSelectOption will be a map[string]string
func (options StringSelectOptionSlice) FindValue(text string) string {
	for _, v := range options {
		if text == v.Text {
			return v.Value
		}
	}
	return text
}

// StringSelectMulti contains properties for multiple select inputs.
type StringSelectMulti struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Options is an array of possible values.
	Options StringSelectOptionSlice `json:"options"`

	// IsRequired is a bool indicating wether an input is needed.
	IsRequired bool `json:"isRequired"`
}

func appendIfNotSelected(value string, input, output []string) []string {
	for _, val := range input {
		if val == value {
			return output
		}
	}
	return append(output, value)
}

// Run implements github.com/stratumn/sdk/generator.Input.
func (in *StringSelectMulti) Run() (interface{}, error) {
	values := make([]string, 0)
	for {
		options := make([]string, 0)
		if in.Default != "" {
			options = appendIfNotSelected(in.Options.FindText(in.Default), values, options)
		}
		options = append(options, "")
		for _, v := range in.Options {
			if in.Default == "" || v.Value != in.Default {
				options = appendIfNotSelected(v.Text, values, options)
			}
		}
		prompt := promptui.Select{
			Label: in.Prompt,
			Items: options,
			Size:  len(options),
		}
		prompt.Templates = new(promptui.SelectTemplates)
		prompt.Templates.Help = fmt.Sprintf(`{{ "Use the arrow keys to navigate:" | faint }} {{ .NextKey | faint }} ` +
			`{{ .PrevKey | faint }} {{ .PageDownKey | faint }} {{ .PageUpKey | faint }} ` +
			`{{ "(select empty line to finish your selection) "| faint }}`)
		_, val, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		if val == "" {
			break
		}
		values = append(values, val)
	}
	if in.IsRequired && len(values) == 0 {
		return nil, errors.New("Selection is mandatory")
	}
	for i, out := range values {
		values[i] = in.Options.FindValue(out)
	}
	return values, nil
}

// StringSlice contains properties for string inputs.
type StringSlice struct {
	InputShared

	// Default is the default value.
	Default string `json:"default"`

	// Format is a string containing a regexp the value must have.
	Format string `json:"format"`
}

// Run implements github.com/stratumn/sdk/generator.Input.
func (in *StringSlice) Run() (interface{}, error) {
	values := make([]interface{}, 0)
	label := fmt.Sprintf("%s %s",
		in.Prompt,
		promptui.Styler(promptui.FGFaint)("(one per line, empty line to finish)"))
	for {
		prompt := in.createStringPrompt(label, fmt.Sprintf("|%s", in.Format), in.Default)
		val, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		if val == "" {
			break
		}
		values = append(values, val)
	}
	if len(values) == 0 {
		return nil, errors.New("list must be non empty")
	}
	return values, nil
}
