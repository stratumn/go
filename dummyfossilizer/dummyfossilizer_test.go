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

package dummyfossilizer

import (
	"testing"

	"github.com/stratumn/sdk/fossilizer"
)

func TestGetInfo(t *testing.T) {
	a := New(&Config{})
	got, err := a.GetInfo()
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	if _, ok := got.(*Info); !ok {
		t.Errorf("a.GetInfo(): info = %#v want *Info", got)
	}
}

func TestFossilize(t *testing.T) {
	a := New(&Config{})
	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

	var (
		data = []byte("data")
		meta = []byte("meta")
	)

	go func() {
		if err := a.Fossilize(data, meta); err != nil {
			t.Errorf("a.Fossilize(): err: %s", err)
		}
	}()

	r := <-rc

	if got, want := string(r.Data), string(data); got != want {
		t.Errorf("<-rc: Data = %q want %q", got, want)
	}
	if got, want := string(r.Meta), string(meta); got != want {
		t.Errorf("<-rc: Meta = %q want %q", got, want)
	}
	if got, want := r.Evidence.(map[string]interface{})["authority"].(string), "dummy"; got != want {
		t.Errorf(`<-rc: Evidence["authority"] = %q want %q`, got, want)
	}
}
