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

package bcbatchfossilizer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/sdk/batchfossilizer"
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

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest) (results []*fossilizer.Result) {
	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := a.Start(ctx); err != nil && errors.Cause(err) != context.Canceled {
			t.Errorf("a.Start(): err: %s", err)
		}
	}()

	<-a.Started()

	for _, test := range tests {
		if err := a.Fossilize(test.data, test.meta); err != nil {
			t.Errorf("a.Fossilize(): err: %s", err)
		}
		if test.sleep > 0 {
			time.Sleep(test.sleep)
		}
	}

RESULT_LOOP:
	for range tests {
		e := <-ec
		r := e.Data.(*fossilizer.Result)
		for i := range tests {
			test := &tests[i]
			if string(test.meta) == string(r.Meta) {
				test.fossilized = true
				if !reflect.DeepEqual(r.Data, test.data) {
					a := fmt.Sprintf("%x", r.Data)
					e := fmt.Sprintf("%x", test.data)
					t.Errorf("test#%d: Data = %q want %q", i, a, e)
				}
				evidence := r.Evidence.Proof.(*evidences.BcBatchProof)
				if !reflect.DeepEqual(evidence.Batch.Path, test.path) {
					ajs, _ := json.MarshalIndent(evidence.Batch.Path, "", "  ")
					ejs, _ := json.MarshalIndent(test.path, "", "  ")
					t.Errorf("test#%d: Path = %s\nwant %s", i, ajs, ejs)
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

	return results
}

func benchmarkFossilize(b *testing.B, config *Config, batchConfig *batchfossilizer.Config) {
	n := b.N

	a, err := New(config, batchConfig)
	if err != nil {
		b.Fatalf("New(): err: %s", err)
	}

	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := a.Start(ctx); err != nil && errors.Cause(err) != context.Canceled {
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
		cancel()
	}()

	for i := 0; i < n; i++ {
		<-ec
	}

	b.StopTimer()
}

func atos(a types.Bytes32) []byte {
	return a[:]
}
