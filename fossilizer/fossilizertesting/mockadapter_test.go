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

package fossilizertesting

import (
	"reflect"
	"testing"

	"github.com/stratumn/sdk/fossilizer"
)

func TestMockAdapter_GetInfo(t *testing.T) {
	a := &MockAdapter{}

	if _, err := a.GetInfo(); err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}

	a.MockGetInfo.Fn = func() (interface{}, error) { return map[string]string{"name": "test"}, nil }
	info, err := a.GetInfo()
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}

	if got, want := info.(map[string]string)["name"], "test"; got != want {
		t.Errorf(`a.GetInfo(): info["name"] = %q want %q`, got, want)
	}
	if got, want := a.MockGetInfo.CalledCount, 2; got != want {
		t.Errorf(`a.MockGetInfo.CalledCount = %d want %d`, got, want)
	}
}

func TestMockAdapter_AddResultChan(t *testing.T) {
	a := &MockAdapter{}

	c1 := make(chan *fossilizer.Result)
	a.AddResultChan(c1)

	a.MockAddResultChan.Fn = func(chan *fossilizer.Result) {}

	c2 := make(chan *fossilizer.Result)
	a.AddResultChan(c2)

	if got, want := a.MockAddResultChan.CalledCount, 2; got != want {
		t.Errorf(`a.MockAddResultChan.CalledCount = %d want %d`, got, want)
	}
	var (
		got  = a.MockAddResultChan.CalledWith
		want = []chan *fossilizer.Result{c1, c2}
	)
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockAddResultChan.CalledWith = %#v want %#v`, got, want)
	}
	if got, want := a.MockAddResultChan.LastCalledWith, c2; got != want {
		t.Errorf(`a.MockAddResultChan.LastCalledWith = %#v want %#v`, got, want)
	}
}

func TestMockAdapter_Fossilize(t *testing.T) {
	a := &MockAdapter{}

	d1 := []byte("data1")
	m1 := []byte("meta1")

	if err := a.Fossilize(d1, m1); err != nil {
		t.Fatalf("a.Fossilize(): err: %s", err)
	}

	a.MockFossilize.Fn = func([]byte, []byte) error { return nil }

	d2 := []byte("data2")
	m2 := []byte("meta2")

	if err := a.Fossilize(d2, m2); err != nil {
		t.Errorf("a.Fossilize(): err: %s", err)
	}

	if got, want := a.MockFossilize.CalledCount, 2; got != want {
		t.Errorf(`a.MockFossilize.CalledCount = %d want %d`, got, want)
	}

	var got []string
	for _, b := range a.MockFossilize.CalledWithData {
		got = append(got, string(b))
	}
	want := []string{string(d1), string(d2)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockFossilize.CalledWithData = %q want %q`, got, want)
	}

	if got, want := string(a.MockFossilize.LastCalledWithData), string(d2); got != want {
		t.Errorf(`a.MockFossilize.LastCalledWithData = %q want %q`, got, want)
	}

	got = nil
	for _, b := range a.MockFossilize.CalledWithMeta {
		got = append(got, string(b))
	}
	want = []string{string(m1), string(m2)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`a.MockFossilize.CalledWithMeta = %q want %q`, got, want)
	}

	if got, want := string(a.MockFossilize.LastCalledWithMeta), string(m2); got != want {
		t.Errorf(`a.MockFossilize.LastCalledWithMeta = %q want %q`, got, want)
	}
}
