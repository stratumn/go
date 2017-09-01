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
	"reflect"
	"testing"
)

func TestInputSliceUnmarshalJSON(t *testing.T) {
	f, err := os.Open("testdata/nodejs/generator.json")
	if err != nil {
		t.Fatalf("err: os.Open(): %s", err)
	}
	dec := json.NewDecoder(f)
	var gen Definition
	if err := dec.Decode(&gen); err != nil {
		t.Fatalf("err: dec.Decode(): %s", err)
	}
	if got, want := gen.Name, "nodejs"; got != want {
		t.Errorf("err: gen.Name: got %q want %q", got, want)
	}
	if got, want := gen.Inputs["name"].Msg(), "Project name: (default \"{{.dir}}\")\n"; got != want {
		t.Errorf(`err: gen.Inputs["name"].Msg(): got %q want %q`, got, want)
	}
}

func TestInputSliceUnmarshalJSON_invalid(t *testing.T) {
	var gen Definition
	if err := json.Unmarshal([]byte(`{"inputs": [1, 2, 3]}`), &gen); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestInputSliceUnmarshalJSON_invalidInput(t *testing.T) {
	var gen Definition
	if err := json.Unmarshal([]byte(`{"inputs": {"test": 1}}`), &gen); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestInputSliceUnmarshalJSON_invalidType(t *testing.T) {
	var gen Definition
	if err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "nope"}}}`), &gen); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestInputSliceUnmarshalJSON_invalidString(t *testing.T) {
	var gen Definition
	if err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "string", "default": 1}}}`), &gen); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestInputSliceUnmarshalJSON_invalidSelect(t *testing.T) {
	var gen Definition
	if err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "select:string", "options": [1]}}}`), &gen); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestInputSliceUnmarshalJSON_invalidSlice(t *testing.T) {
	var gen Definition
	if err := json.Unmarshal([]byte(`{"inputs": {"test": {"type": "slice:string", "separator": 1}}}`), &gen); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringInputSet(t *testing.T) {
	in := StringInput{}
	if err := in.Set("hello"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := in.value, "hello"; got != want {
		t.Errorf("err: in.value: got %q want %q", got, want)
	}
}

func TestStringInputSet_default(t *testing.T) {
	in := StringInput{Default: "hello"}
	if err := in.Set(""); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := in.value, "hello"; got != want {
		t.Errorf("err: in.value: got %q want %q", got, want)
	}
}

func TestStringInputSet_notString(t *testing.T) {
	in := StringInput{}
	if err := in.Set(3); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringInputSet_formatValid(t *testing.T) {
	in := StringInput{Format: "[a-z]{4}"}
	if err := in.Set("hello"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := in.value, "hello"; got != want {
		t.Errorf("err: in.value: got %q want %q", got, want)
	}
}

func TestStringInputSet_formatInvalid(t *testing.T) {
	in := StringInput{Format: "[0-9]{4}"}
	if err := in.Set("hello"); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringInputSet_invalidFormat(t *testing.T) {
	in := StringInput{Format: "("}
	if err := in.Set("("); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringInputGet(t *testing.T) {
	in := StringInput{value: "hello", Default: "world"}
	if got, want := in.Get(), "hello"; got != want {
		t.Errorf("err: in.Get(): got %q want %q", got, want)
	}
}

func TestStringInputGet_default(t *testing.T) {
	in := StringInput{Default: "hello"}
	if got, want := in.Get(), "hello"; got != want {
		t.Errorf("err: in.Get(): got %q want %q", got, want)
	}
}

func TestStringInputMsg(t *testing.T) {
	in := StringInput{InputShared: InputShared{Prompt: "what:"}}
	if got, want := in.Msg(), "what:\n"; got != want {
		t.Errorf("err: in.Msg(): got %q want %q", got, want)
	}
}

func TestStringInputMsg_default(t *testing.T) {
	in := StringInput{InputShared: InputShared{Prompt: "what:"}, Default: "nothing"}
	if got, want := in.Msg(), "what: (default \"nothing\")\n"; got != want {
		t.Errorf("err: in.Msg(): got %q want %q", got, want)
	}
}

func TestStringSelectSet(t *testing.T) {
	in := StringSelect{
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
	}
	if err := in.Set("y"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := in.value, "y"; got != want {
		t.Errorf("err: in.value: got %q want %q", got, want)
	}
}

func TestStringSelectSet_default(t *testing.T) {
	in := StringSelect{
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
		Default: "y",
	}
	if err := in.Set(""); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := in.value, "y"; got != want {
		t.Errorf("err: in.value: got %q want %q", got, want)
	}
}

func TestStringSelectSet_notString(t *testing.T) {
	in := StringSelect{
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
	}
	if err := in.Set(3); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSelectSet_invalid(t *testing.T) {
	in := StringSelect{
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
	}
	if err := in.Set("hello"); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSelectGet(t *testing.T) {
	in := StringSelect{
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
		Default: "y",
		value:   "n",
	}
	if got, want := in.Get(), "n"; got != want {
		t.Errorf("err: in.Get(): got %q want %q", got, want)
	}
}

func TestStringSelectGet_default(t *testing.T) {
	in := StringSelect{
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
		Default: "y",
	}
	if got, want := in.Get(), "y"; got != want {
		t.Errorf("err: in.Get(): got %q want %q", got, want)
	}
}

func TestStringSelectMsg(t *testing.T) {
	in := StringSelect{
		InputShared: InputShared{
			Prompt: "value:",
		},
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
	}
	want := `value:
y: yes
n: no
`
	if got := in.Msg(); got != want {
		t.Errorf("err: in.Msg(): got %q want %q", got, want)
	}
}

func TestStringSelectMsg_default(t *testing.T) {
	in := StringSelect{
		InputShared: InputShared{
			Prompt: "value:",
		},
		Options: []StringSelectOption{
			StringSelectOption{
				Input: "y",
				Value: "y",
				Text:  "yes",
			},
			StringSelectOption{
				Input: "n",
				Value: "n",
				Text:  "no",
			},
		},
		Default: "n",
	}
	want := `value:
y: yes
n: no (default)
`
	if got := in.Msg(); got != want {
		t.Errorf("err: in.Msg(): got %q want %q", got, want)
	}
}

func TestStringSliceSet_one(t *testing.T) {
	in := StringSlice{Separator: ","}
	if err := in.Set("hello"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 1; got != want {
		t.Fatalf("err: len(in.values): got %d want %d", got, want)
	}
	if got, want := in.values, []string{"hello"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_two(t *testing.T) {
	in := StringSlice{Separator: ","}
	if err := in.Set("hello,hi"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 2; got != want {
		t.Errorf("err: len(in.values): got %q want %q", got, want)
	}
	if got, want := in.values, []string{"hello", "hi"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_defaultOne(t *testing.T) {
	in := StringSlice{Separator: ",", Default: "hello"}
	if err := in.Set(""); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 1; got != want {
		t.Fatalf("err: len(in.values): got %d want %d", got, want)
	}
	if got, want := in.values, []string{"hello"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_defaultTwo(t *testing.T) {
	in := StringSlice{Separator: ",", Default: "hello,hi"}
	if err := in.Set(""); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 2; got != want {
		t.Errorf("err: len(in.values): got %q want %q", got, want)
	}
	if got, want := in.values, []string{"hello", "hi"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_emptyList(t *testing.T) {
	in := StringSlice{}
	if err := in.Set(""); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSliceSet_notString(t *testing.T) {
	in := StringSlice{}
	if err := in.Set(3); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSliceSet_withoutSeparator(t *testing.T) {
	in := StringSlice{}
	if err := in.Set("hello"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 5; got != want {
		t.Fatalf("err: len(in.values): got %d want %d", got, want)
	}
	if got, want := in.values, []string{"h", "e", "l", "l", "o"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_withSemicolonSeparator(t *testing.T) {
	in := StringSlice{Separator: ";"}
	if err := in.Set("Bit,Coin;ether/eum;Tender%mint"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 3; got != want {
		t.Fatalf("err: len(in.values): got %d want %d", got, want)
	}
	if got, want := in.values, []string{"Bit,Coin", "ether/eum", "Tender%mint"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_formatValid(t *testing.T) {
	in := StringSlice{Separator: ";", Format: "^[a-z].*"}
	if err := in.Set("bitcoin;ethereum;tendermint"); err != nil {
		t.Fatalf("err: in.Set(): %s", err)
	}
	if got, want := len(in.values), 3; got != want {
		t.Fatalf("err: len(in.values): got %d want %d", got, want)
	}
	if got, want := in.values, []string{"bitcoin", "ethereum", "tendermint"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.values: got %v want %v", got, want)
	}
}

func TestStringSliceSet_formatInvalidFirst(t *testing.T) {
	in := StringSlice{Separator: ";", Format: "^[a-z].*"}
	if err := in.Set("Bitcoin;ethereum;tendermint"); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSliceSet_formatInvalidOther(t *testing.T) {
	in := StringSlice{Separator: ";", Format: "^[a-z].*"}
	if err := in.Set("bitcoin;ethereum;Tendermint"); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSliceSet_invalidFormat(t *testing.T) {
	in := StringSlice{Format: "("}
	if err := in.Set("("); err == nil {
		t.Error("err: err = nil want Error")
	}
}

func TestStringSliceGet(t *testing.T) {
	in := StringSlice{values: []string{"bitcoin", "ethereum", "tendermint"}, Default: "hello,world", Separator: ","}
	if got, want := in.Get(), []string{"bitcoin", "ethereum", "tendermint"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.Get(): got %v want %v", got, want)
	}
}

func TestStringSliceGet_default(t *testing.T) {
	in := StringSlice{Default: "hello,world", Separator: ","}
	if got, want := in.Get(), []string{"hello", "world"}; !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.Get(): got %v want %v", got, want)
	}
}

func TestStringSliceGet_defaultWithoutSeparator(t *testing.T) {
	in := StringSlice{Default: "hello,world"}
	if got, want := in.Get(), []string(nil); !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.Get(): got %#v want %#v", got, want)
	}
}

func TestStringSliceGet_defaultWithoutDefault(t *testing.T) {
	in := StringSlice{Default: noValue, Separator: ","}
	if got, want := in.Get(), []string(nil); !reflect.DeepEqual(got, want) {
		t.Errorf("err: in.Get(): got %#v want %#v", got, want)
	}
}

func TestStringSliceMsg(t *testing.T) {
	in := StringSlice{
		values:      []string{"bitcoin", "ethereum", "tendermint"},
		Separator:   ";",
		InputShared: InputShared{Prompt: "what:"},
	}
	if got, want := in.Msg(), "what: (separator \";\")\n"; got != want {
		t.Errorf("err: in.Msg(): got %q want %q", got, want)
	}
}

func TestStringSliceMsg_default(t *testing.T) {
	in := StringSlice{
		values:      []string{"bitcoin", "ethereum", "tendermint"},
		Default:     "hello,world",
		Separator:   ";",
		InputShared: InputShared{Prompt: "what:"},
	}
	if got, want := in.Msg(), "what: (separator \";\") (default \"hello,world\")\n"; got != want {
		t.Errorf("err: in.Msg(): got %q want %q", got, want)
	}
}
