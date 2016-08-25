// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/testutil"
	"github.com/stratumn/goprivate/types"
)

type fossilizeTest struct {
	data       []byte
	meta       []byte
	path       merkle.Path
	sleep      time.Duration
	fossilized bool
}

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest, start bool, fossilize bool) {
	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

	if start {
		go func() {
			if err := a.Start(); err != nil {
				t.Fatal(err)
			}
		}()

		defer func() {
			a.Stop()
			close(rc)
		}()
	}

	if fossilize {
		for _, test := range tests {
			if err := a.Fossilize(test.data, test.meta); err != nil {
				t.Fatal(err)
			}
			if test.sleep > 0 {
				time.Sleep(test.sleep)
			}
		}
	}

RESULT_LOOP:
	for _ = range tests {
		r := <-rc
		for i := range tests {
			test := &tests[i]
			if string(test.meta) == string(r.Meta) {
				test.fossilized = true
				if !reflect.DeepEqual(r.Data, test.data) {
					a := hex.EncodeToString(r.Data)
					e := hex.EncodeToString(test.data)
					t.Logf("actual: %s; expected %s\n", a, e)
					t.Error("unexpected result data")
				}
				evidence := r.Evidence.(*EvidenceWrapper).Evidence
				if !reflect.DeepEqual(evidence.Path, test.path) {
					ajs, _ := json.MarshalIndent(evidence.Path, "", "  ")
					ejs, _ := json.MarshalIndent(test.path, "", "  ")
					t.Logf("actual: %s; expected %s\n", string(ajs), string(ejs))
					t.Error("unexpected merkle path")
				}
				continue RESULT_LOOP
			}
		}
		t.Errorf("unexpected result meta: %s", r.Meta)
	}

	for _, test := range tests {
		if !test.fossilized {
			t.Errorf("not fossilized: %s\n", test.meta)
		}
	}
}

func benchmarkFossilize(b *testing.B, config *Config) {
	a, err := New(config)
	if err != nil {
		b.Fatal(err)
	}
	go func() {
		if err := a.Start(); err != nil {
			b.Fatal(err)
		}
	}()
	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)
	defer func() {
		a.Stop()
		close(rc)
	}()

	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = atos(*testutil.RandomHash())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := a.Fossilize(data[i], data[i]); err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; i < b.N; i++ {
		<-rc
	}
}

func atos(a types.Bytes32) []byte {
	return a[:]
}
