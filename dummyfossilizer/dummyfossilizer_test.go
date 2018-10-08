// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-core/dummyfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	a := New(&Config{})
	got, err := a.GetInfo(context.Background())
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
		if err := a.Fossilize(context.Background(), data, meta); err != nil {
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
		err := a.Fossilize(context.Background(), data, meta)
		assert.NoError(t, err, "a.Fossilize()")
	}()

	e := <-ec
	r := e.Data.(*fossilizer.Result)

	t.Run("Time()", func(t *testing.T) {
		p, err := evidences.UnmarshalProof(&r.Evidence)
		require.NoError(t, err)

		assert.Equal(t, timestamp, p.Time())
	})

	t.Run("Verify()", func(t *testing.T) {
		p, err := evidences.UnmarshalProof(&r.Evidence)
		require.NoError(t, err)

		assert.True(t, p.Verify(""))
	})
}
