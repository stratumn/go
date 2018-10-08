// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/batchfossilizer"
	"github.com/stratumn/go-core/bcbatchfossilizer/evidences"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/types"
	mktypes "github.com/stratumn/merkle/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fossilizeTest struct {
	data       []byte
	meta       []byte
	path       mktypes.Path
	sleep      time.Duration
	fossilized bool
}

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest) (results []*fossilizer.Result) {
	ec := make(chan *fossilizer.Event, 1)
	a.AddFossilizerEventChan(ec)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := a.Start(ctx); err != nil {
			assert.EqualError(t, errors.Cause(err), context.Canceled.Error())
		}
	}()

	<-a.Started()

	for _, test := range tests {
		err := a.Fossilize(context.Background(), test.data, test.meta)
		assert.NoError(t, err, "a.Fossilize()")

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
				assert.Equal(t, test.data, r.Data)

				proof, err := evidences.UnmarshalProof(&r.Evidence)
				require.NoError(t, err)
				assert.Equal(t, test.path, proof.Batch.Path)

				results = append(results, r)
				continue RESULT_LOOP
			}
		}
		a := fmt.Sprintf("%x", r.Meta)
		t.Errorf("unexpected Meta %q", a)
	}

	for _, test := range tests {
		assert.True(t, test.fossilized)
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
		data[i] = chainscripttest.RandomHash()
	}

	<-a.Started()

	b.ResetTimer()
	log.SetOutput(ioutil.Discard)

	go func() {
		for i := 0; i < n; i++ {
			if err := a.Fossilize(context.Background(), data[i], data[i]); err != nil {
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
