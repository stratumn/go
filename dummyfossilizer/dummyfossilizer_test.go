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
	"bytes"
	"fmt"
	"testing"
	"time"

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
	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)

	var (
		data = []byte("data")
		meta = []byte("meta")
	)

	go func() {
		if err := a.Fossilize(data, meta); err != nil {
			t.Errorf("a.Fossilize(): err: %s", err)
		}
	}()

	e := <-ec
	r := e.Data.(*fossilizer.Result)

	if got, want := string(r.Data), string(data); got != want {
		t.Errorf("<-rc: Data = %q want %q", got, want)
	}
	if got, want := string(r.Meta), string(meta); got != want {
		t.Errorf("<-rc: Meta = %q want %q", got, want)
	}
	if got, want := r.Evidence.Provider, "dummy"; got != want {
		t.Errorf(`<-rc: Evidence.Provider = %s want %s`, got, want)
	}
}

func TestDummyProof(t *testing.T) {
	a := New(&Config{})
	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)

	var (
		data      = []byte("data")
		meta      = []byte("meta")
		timestamp = uint64(time.Now().Unix())
	)

	go func() {
		if err := a.Fossilize(data, meta); err != nil {
			t.Errorf("a.Fossilize(): err: %s", err)
		}
	}()

	e := <-ec
	r := e.Data.(*fossilizer.Result)

	t.Run("Time()", func(t *testing.T) {
		if got, want := r.Evidence.Proof, timestamp; got.Time() != want {
			t.Errorf(`<-rc: Evidence.originalProof.Time() = %d, want %d`, got.Time(), want)
		}
	})

	t.Run("FullProof()", func(t *testing.T) {
		got := r.Evidence.Proof
		want := fmt.Sprintf("{\n   \"timestamp\": %d\n}", got.Time())
		if bytes.Compare(got.FullProof(), []byte(want)) != 0 {
			t.Errorf(`<-rc: Evidence.originalProof.FullProof() = %s, want %s`, got.FullProof(), want)
		}
	})

	t.Run("Verify()", func(t *testing.T) {
		if got, want := r.Evidence.Proof.Verify(""), true; got != want {
			t.Errorf(`<-rc: Evidence.originalProof.Verify() = %v, want %v`, got, want)
		}
	})
}
