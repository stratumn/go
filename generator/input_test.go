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

// func TestStringInput(t *testing.T) {
// 	t.Run("Success", func(t *testing.T) {
// 		in := StringInput{}
// 		in.SetStdin(strings.NewReader("lol"))
// 		val, err := in.Run()
// 		assert.NoError(t, err)
// 		assert.IsType(t, val, "lol")
// 	})
// }

// func TestStringInputSet(t *testing.T) {
// 	in := StringInput{}
// 	if err := in.Set("hello"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.value, "hello"; got != want {
// 		t.Errorf("err: in.value: got %q want %q", got, want)
// 	}
// }

// func TestStringInputSet_default(t *testing.T) {
// 	in := StringInput{Default: "hello"}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.value, "hello"; got != want {
// 		t.Errorf("err: in.value: got %q want %q", got, want)
// 	}
// }

// func TestStringInputSet_notString(t *testing.T) {
// 	in := StringInput{}
// 	if err := in.Set(3); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringInputSet_formatValid(t *testing.T) {
// 	in := StringInput{Format: "[a-z]{4}"}
// 	if err := in.Set("hello"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.value, "hello"; got != want {
// 		t.Errorf("err: in.value: got %q want %q", got, want)
// 	}
// }

// func TestStringInputSet_formatInvalid(t *testing.T) {
// 	in := StringInput{Format: "[0-9]{4}"}
// 	if err := in.Set("hello"); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringInputSet_invalidFormat(t *testing.T) {
// 	in := StringInput{Format: "("}
// 	if err := in.Set("("); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringInputGet(t *testing.T) {
// 	in := StringInput{value: "hello", Default: "world"}
// 	if got, want := in.Get(), "hello"; got != want {
// 		t.Errorf("err: in.Get(): got %q want %q", got, want)
// 	}
// }

// func TestStringInputGet_default(t *testing.T) {
// 	in := StringInput{Default: "hello"}
// 	if got, want := in.Get(), "hello"; got != want {
// 		t.Errorf("err: in.Get(): got %q want %q", got, want)
// 	}
// }

// func TestStringInputMsg(t *testing.T) {
// 	in := StringInput{InputShared: InputShared{Prompt: "what:"}}
// 	if got, want := in.Msg(), "what:\n"; got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringInputMsg_default(t *testing.T) {
// 	in := StringInput{InputShared: InputShared{Prompt: "what:"}, Default: "nothing"}
// 	if got, want := in.Msg(), "what: (default \"nothing\")\n"; got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectSet(t *testing.T) {
// 	in := StringSelect{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set("y"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.value, "y"; got != want {
// 		t.Errorf("err: in.value: got %q want %q", got, want)
// 	}
// }

// func TestStringSelectSet_default(t *testing.T) {
// 	in := StringSelect{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "y",
// 	}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.value, "y"; got != want {
// 		t.Errorf("err: in.value: got %q want %q", got, want)
// 	}
// }

// func TestStringSelectSet_notString(t *testing.T) {
// 	in := StringSelect{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set(3); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSelectSet_invalid(t *testing.T) {
// 	in := StringSelect{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set("hello"); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSelectGet(t *testing.T) {
// 	in := StringSelect{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "y",
// 		value:   "n",
// 	}
// 	if got, want := in.Get(), "n"; got != want {
// 		t.Errorf("err: in.Get(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectGet_default(t *testing.T) {
// 	in := StringSelect{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "y",
// 	}
// 	if got, want := in.Get(), "y"; got != want {
// 		t.Errorf("err: in.Get(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMsg(t *testing.T) {
// 	in := StringSelect{
// 		InputShared: InputShared{
// 			Prompt: "value:",
// 		},
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	want := `value:
// y: yes
// n: no
// `
// 	if got := in.Msg(); got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMsg_default(t *testing.T) {
// 	in := StringSelect{
// 		InputShared: InputShared{
// 			Prompt: "value:",
// 		},
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "n",
// 	}
// 	want := `value:
// y: yes
// n: no (default)
// `
// 	if got := in.Msg(); got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiSet(t *testing.T) {
// 	in := StringSelectMulti{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set("y"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.values, []string{"y"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %q want %q", got, want)
// 	}
// }
// func TestStringSelectMultiSet_multi(t *testing.T) {
// 	in := StringSelectMulti{
// 		Separator: ",",
// 		Options: []StringSelectOption{
// 			{
// 				Input: "1",
// 				Value: "1",
// 				Text:  "one",
// 			},
// 			{
// 				Input: "2",
// 				Value: "2",
// 				Text:  "two",
// 			},
// 		},
// 	}
// 	if err := in.Set("1,2"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.values, []string{"1", "2"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiSet_default(t *testing.T) {
// 	in := StringSelectMulti{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "y",
// 	}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.values, []string{"y"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiSet_notString(t *testing.T) {
// 	in := StringSelectMulti{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set(3); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSelectMultiSet_invalid(t *testing.T) {
// 	in := StringSelectMulti{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set("hello"); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSelectMultiSet_requiredNoDefault(t *testing.T) {
// 	in := StringSelectMulti{
// 		IsRequired: true,
// 		Separator:  ",",
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set(""); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSelectMultiSet_requiredDefault(t *testing.T) {
// 	in := StringSelectMulti{
// 		IsRequired: true,
// 		Default:    "1,2",
// 		Separator:  ",",
// 		Options: []StringSelectOption{
// 			{
// 				Input: "1",
// 				Value: "1",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "2",
// 				Value: "2",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := in.values, []string{"1", "2"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiSet_noSeparator(t *testing.T) {
// 	in := StringSelectMulti{
// 		IsRequired: true,
// 		Options: []StringSelectOption{
// 			{
// 				Input: "1",
// 				Value: "1",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "2",
// 				Value: "2",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set("1,2"); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSelectMultiSet_notRequired(t *testing.T) {
// 	in := StringSelectMulti{
// 		IsRequired: false,
// 		Separator:  ",",
// 		Options: []StringSelectOption{
// 			{
// 				Input: "1",
// 				Value: "1",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "2",
// 				Value: "2",
// 				Text:  "no",
// 			},
// 		},
// 	}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 0; got != want {
// 		t.Errorf("err: len(in.values): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiGet(t *testing.T) {
// 	in := StringSelectMulti{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "y",
// 		values:  []string{"n"},
// 	}
// 	if got, want := in.Get(), []string{"n"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.Get(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiGet_default(t *testing.T) {
// 	in := StringSelectMulti{
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Separator: ",",
// 		Default:   "y,n",
// 	}
// 	if got, want := in.Get(), []string{"y", "n"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.Get(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiMsg(t *testing.T) {
// 	in := StringSelectMulti{
// 		InputShared: InputShared{
// 			Prompt: "value:",
// 		},
// 		Separator: ",",
// 		Options: []StringSelectOption{
// 			{
// 				Input: "1",
// 				Value: "1",
// 				Text:  "one",
// 			},
// 			{
// 				Input: "2",
// 				Value: "2",
// 				Text:  "two",
// 			},
// 		},
// 	}
// 	want := `value: (separator ",") (default: None)
// 1: one
// 2: two
// `
// 	if got := in.Msg(); got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }
// func TestStringSelectMultiMsg_required(t *testing.T) {
// 	in := StringSelectMulti{
// 		InputShared: InputShared{
// 			Prompt: "value:",
// 		},
// 		Separator:  ",",
// 		IsRequired: true,
// 		Options: []StringSelectOption{
// 			{
// 				Input: "1",
// 				Value: "1",
// 				Text:  "one",
// 			},
// 			{
// 				Input: "2",
// 				Value: "2",
// 				Text:  "two",
// 			},
// 		},
// 	}
// 	want := `value: (separator ",")
// 1: one
// 2: two
// `
// 	if got := in.Msg(); got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringSelectMultiMsg_default(t *testing.T) {
// 	in := StringSelectMulti{
// 		InputShared: InputShared{
// 			Prompt: "value:",
// 		},
// 		Options: []StringSelectOption{
// 			{
// 				Input: "y",
// 				Value: "y",
// 				Text:  "yes",
// 			},
// 			{
// 				Input: "n",
// 				Value: "n",
// 				Text:  "no",
// 			},
// 		},
// 		Default: "n",
// 	}
// 	want := `value: (separator "")
// y: yes
// n: no (default)
// `
// 	if got := in.Msg(); got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringSliceSet_one(t *testing.T) {
// 	in := StringSlice{Separator: ","}
// 	if err := in.Set("hello"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 1; got != want {
// 		t.Fatalf("err: len(in.values): got %d want %d", got, want)
// 	}
// 	if got, want := in.values, []string{"hello"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_two(t *testing.T) {
// 	in := StringSlice{Separator: ","}
// 	if err := in.Set("hello,hi"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 2; got != want {
// 		t.Errorf("err: len(in.values): got %q want %q", got, want)
// 	}
// 	if got, want := in.values, []string{"hello", "hi"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_defaultOne(t *testing.T) {
// 	in := StringSlice{Separator: ",", Default: "hello"}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 1; got != want {
// 		t.Fatalf("err: len(in.values): got %d want %d", got, want)
// 	}
// 	if got, want := in.values, []string{"hello"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_defaultTwo(t *testing.T) {
// 	in := StringSlice{Separator: ",", Default: "hello,hi"}
// 	if err := in.Set(""); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 2; got != want {
// 		t.Errorf("err: len(in.values): got %q want %q", got, want)
// 	}
// 	if got, want := in.values, []string{"hello", "hi"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_emptyList(t *testing.T) {
// 	in := StringSlice{}
// 	if err := in.Set(""); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSliceSet_notString(t *testing.T) {
// 	in := StringSlice{}
// 	if err := in.Set(3); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSliceSet_withoutSeparator(t *testing.T) {
// 	in := StringSlice{}
// 	if err := in.Set("hello"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 5; got != want {
// 		t.Fatalf("err: len(in.values): got %d want %d", got, want)
// 	}
// 	if got, want := in.values, []string{"h", "e", "l", "l", "o"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_withSemicolonSeparator(t *testing.T) {
// 	in := StringSlice{Separator: ";"}
// 	if err := in.Set("Bit,Coin;ether/eum;Tender%mint"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 3; got != want {
// 		t.Fatalf("err: len(in.values): got %d want %d", got, want)
// 	}
// 	if got, want := in.values, []string{"Bit,Coin", "ether/eum", "Tender%mint"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_formatValid(t *testing.T) {
// 	in := StringSlice{Separator: ";", Format: "^[a-z].*"}
// 	if err := in.Set("bitcoin;ethereum;tendermint"); err != nil {
// 		t.Fatalf("err: in.Set(): %s", err)
// 	}
// 	if got, want := len(in.values), 3; got != want {
// 		t.Fatalf("err: len(in.values): got %d want %d", got, want)
// 	}
// 	if got, want := in.values, []string{"bitcoin", "ethereum", "tendermint"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.values: got %v want %v", got, want)
// 	}
// }

// func TestStringSliceSet_formatInvalidFirst(t *testing.T) {
// 	in := StringSlice{Separator: ";", Format: "^[a-z].*"}
// 	if err := in.Set("Bitcoin;ethereum;tendermint"); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSliceSet_formatInvalidOther(t *testing.T) {
// 	in := StringSlice{Separator: ";", Format: "^[a-z].*"}
// 	if err := in.Set("bitcoin;ethereum;Tendermint"); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSliceSet_invalidFormat(t *testing.T) {
// 	in := StringSlice{Format: "("}
// 	if err := in.Set("("); err == nil {
// 		t.Error("err: err = nil want Error")
// 	}
// }

// func TestStringSliceGet(t *testing.T) {
// 	in := StringSlice{values: []string{"bitcoin", "ethereum", "tendermint"}, Default: "hello,world", Separator: ","}
// 	if got, want := in.Get(), []string{"bitcoin", "ethereum", "tendermint"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.Get(): got %v want %v", got, want)
// 	}
// }

// func TestStringSliceGet_default(t *testing.T) {
// 	in := StringSlice{Default: "hello,world", Separator: ","}
// 	if got, want := in.Get(), []string{"hello", "world"}; !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.Get(): got %v want %v", got, want)
// 	}
// }

// func TestStringSliceGet_defaultWithoutSeparator(t *testing.T) {
// 	in := StringSlice{Default: "hello,world"}
// 	if got, want := in.Get(), []string(nil); !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.Get(): got %#v want %#v", got, want)
// 	}
// }

// func TestStringSliceGet_defaultWithoutDefault(t *testing.T) {
// 	in := StringSlice{Default: noValue, Separator: ","}
// 	if got, want := in.Get(), []string(nil); !reflect.DeepEqual(got, want) {
// 		t.Errorf("err: in.Get(): got %#v want %#v", got, want)
// 	}
// }

// func TestStringSliceMsg(t *testing.T) {
// 	in := StringSlice{
// 		values:      []string{"bitcoin", "ethereum", "tendermint"},
// 		Separator:   ";",
// 		InputShared: InputShared{Prompt: "what:"},
// 	}
// 	if got, want := in.Msg(), "what: (separator \";\")\n"; got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }

// func TestStringSliceMsg_default(t *testing.T) {
// 	in := StringSlice{
// 		values:      []string{"bitcoin", "ethereum", "tendermint"},
// 		Default:     "hello,world",
// 		Separator:   ";",
// 		InputShared: InputShared{Prompt: "what:"},
// 	}
// 	if got, want := in.Msg(), "what: (separator \";\") (default \"hello,world\")\n"; got != want {
// 		t.Errorf("err: in.Msg(): got %q want %q", got, want)
// 	}
// }
