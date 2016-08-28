// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package bcbatchfossilizer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/goprivate/batchfossilizer"
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

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest) {
	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

	go func() {
		if err := a.Start(); err != nil {
			t.Error(err)
		}
	}()
	defer func() {
		if err := a.Stop(); err != nil {
			t.Error(err)
		}
		close(rc)
	}()

	for _, test := range tests {
		if err := a.Fossilize(test.data, test.meta); err != nil {
			t.Fatal(err)
		}
		if test.sleep > 0 {
			time.Sleep(test.sleep)
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
					a := fmt.Sprintf("%x", r.Data)
					e := fmt.Sprintf("%x", test.data)
					t.Errorf("test#%d: Data = %q want %q", i, a, e)
				}
				evidence := r.Evidence.(map[string]*Evidence)["dummy"]
				if !reflect.DeepEqual(evidence.Path, test.path) {
					ajs, _ := json.MarshalIndent(evidence.Path, "", "  ")
					ejs, _ := json.MarshalIndent(test.path, "", "  ")
					t.Errorf("test#%d: Path = %s\nwant %s", i, ajs, ejs)
				}
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
}

func benchmarkFossilize(b *testing.B, config *Config, batchConfig *batchfossilizer.Config) {
	a, err := New(config, batchConfig)
	if err != nil {
		b.Fatal(err)
	}

	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

	go func() {
		if err := a.Start(); err != nil {
			b.Error(err)
		}
	}()

	defer func() {
		if err := a.Stop(); err != nil {
			b.Error(err)
		}
		close(rc)
	}()

	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = atos(*testutil.RandomHash())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := a.Fossilize(data[i], data[i]); err != nil {
			b.Error(err)
		}
	}

	for i := 0; i < b.N; i++ {
		<-rc
	}
}

func atos(a types.Bytes32) []byte {
	return a[:]
}
