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

package batchfossilizer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/sdk/cs/evidences"
	"github.com/stratumn/sdk/fossilizer"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

type fossilizeTest struct {
	data       []byte
	meta       []byte
	path       types.Path
	sleep      time.Duration
	fossilized bool
}

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest, start bool, fossilize bool) (results []*fossilizer.Result) {
	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)

	if start {
		go func() {
			if err := a.Start(); err != nil {
				t.Errorf("a.Start(): err: %s", err)
			}
		}()
	}

	<-a.Started()

	if fossilize {
		for _, test := range tests {
			if err := a.Fossilize(test.data, test.meta); err != nil {
				t.Errorf("a.Fossilize(): err: %s", err)
			}
			if test.sleep > 0 {
				time.Sleep(test.sleep)
			}
		}
	}

RESULT_LOOP:
	for range tests {
		e := <-ec
		r := e.Data.(*fossilizer.Result)
		for i := range tests {
			test := &tests[i]
			if fmt.Sprint(test.meta) == fmt.Sprint(r.Meta) {
				test.fossilized = true
				if !reflect.DeepEqual(r.Data, test.data) {
					got := fmt.Sprintf("%x", r.Data)
					want := fmt.Sprintf("%x", test.data)
					t.Errorf("test#%d: Data = %q want %q", i, got, want)
				}
				evidence := r.Evidence.Proof.(*evidences.BatchProof)
				if !reflect.DeepEqual(evidence.Path, test.path) {
					got, _ := json.MarshalIndent(evidence.Path, "", "  ")
					want, _ := json.MarshalIndent(test.path, "", "  ")
					t.Errorf("test#%d: Path = %s\nwant %s", i, got, want)
				}
				results = append(results, r)
				continue RESULT_LOOP
			}
		}
		a := fmt.Sprintf("%x", r.Meta)
		t.Errorf("unexpected Meta %q", a)
	}

	for i, test := range tests {
		if !test.fossilized {
			t.Errorf("test#%d: not fossilized", i)
		}
	}

	if start {
		a.Stop()
	}
	return
}

func benchmarkFossilize(b *testing.B, config *Config) {
	n := b.N

	a, err := New(config)
	if err != nil {
		b.Fatalf("New(): err: %s", err)
	}

	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)

	go func() {
		if err := a.Start(); err != nil {
			b.Errorf("a.Start(): err: %s", err)
		}
	}()

	data := make([][]byte, n)
	for i := 0; i < n; i++ {
		data[i] = atos(*testutil.RandomHash())
	}

	<-a.Started()

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	go func() {
		for i := 0; i < n; i++ {
			if err := a.Fossilize(data[i], data[i]); err != nil {
				b.Errorf("a.Fossilize(): err: %s", err)
			}
		}
		a.Stop()
	}()

	for i := 0; i < n; i++ {
		<-ec
	}

	b.StopTimer()
}

func atos(a types.Bytes32) []byte {
	return a[:]
}
