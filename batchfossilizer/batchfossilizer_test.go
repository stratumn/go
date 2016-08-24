// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stratumn/go/fossilizer"
	"github.com/stratumn/goprivate/merkle"
	"github.com/stratumn/goprivate/testutil"
	"github.com/stratumn/goprivate/types"
)

const interval = 10 * time.Millisecond

var (
	pathA0     merkle.Path
	pathAB0    merkle.Path
	pathAB1    merkle.Path
	pathABC0   merkle.Path
	pathABC1   merkle.Path
	pathABC2   merkle.Path
	pathABCD0  merkle.Path
	pathABCD1  merkle.Path
	pathABCD2  merkle.Path
	pathABCD3  merkle.Path
	pathABCDE0 merkle.Path
	pathABCDE1 merkle.Path
	pathABCDE2 merkle.Path
	pathABCDE3 merkle.Path
	pathABCDE4 merkle.Path
)

func TestMain(m *testing.M) {
	seed := int64(time.Now().Nanosecond())
	fmt.Printf("using seed %d\n", seed)
	rand.Seed(seed)

	loadPath("testdata/path-a-0.json", &pathA0)
	loadPath("testdata/path-ab-0.json", &pathAB0)
	loadPath("testdata/path-ab-1.json", &pathAB1)
	loadPath("testdata/path-abc-0.json", &pathABC0)
	loadPath("testdata/path-abc-1.json", &pathABC1)
	loadPath("testdata/path-abc-2.json", &pathABC2)
	loadPath("testdata/path-abcd-0.json", &pathABCD0)
	loadPath("testdata/path-abcd-1.json", &pathABCD1)
	loadPath("testdata/path-abcd-2.json", &pathABCD2)
	loadPath("testdata/path-abcd-3.json", &pathABCD3)
	loadPath("testdata/path-abcde-0.json", &pathABCDE0)
	loadPath("testdata/path-abcde-1.json", &pathABCDE1)
	loadPath("testdata/path-abcde-2.json", &pathABCDE2)
	loadPath("testdata/path-abcde-3.json", &pathABCDE3)
	loadPath("testdata/path-abcde-4.json", &pathABCDE4)

	flag.Parse()
	os.Exit(m.Run())
}

func TestGetInfo(t *testing.T) {
	a := New(&Config{})
	info, err := a.GetInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("info is nil")
	}
}

func loadPath(filename string, path *merkle.Path) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, path); err != nil {
		panic(err)
	}
}

func TestFossilize(t *testing.T) {
	a := New(&Config{Interval: interval})
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests)
}

func TestFossilizeMaxLeaves(t *testing.T) {
	a := New(&Config{Interval: interval, MaxLeaves: 4})
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 1"), pathABCD0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 1"), pathABCD1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 1"), pathABCD2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d 1"), pathABCD3, 0, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 2"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 2"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 2"), pathABC2, 0, false},
	}
	testFossilizeMultiple(t, a, tests)
}

func TestFossilizeInterval(t *testing.T) {
	a := New(&Config{Interval: interval})
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 1"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 1"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 1"), pathABC2, interval * 2, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 2"), pathABCD0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 2"), pathABCD1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 2"), pathABCD2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d 2"), pathABCD3, interval * 2, false},
		{atos(sha256.Sum256([]byte("a"))), []byte("test a 3"), pathABC0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b 3"), pathABC1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c 3"), pathABC2, 0, false},
	}
	testFossilizeMultiple(t, a, tests)
}

func BenchmarkFossilizeMaxLeaves100(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100})
}

func BenchmarkFossilizeMaxLeaves1000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000})
}

func BenchmarkFossilizeMaxLeaves10000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 10000})
}

func BenchmarkFossilizeMaxLeaves100000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100000})
}

func BenchmarkFossilizeMaxLeaves1000000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000000})
}

type fossilizeTest struct {
	data       []byte
	meta       []byte
	path       merkle.Path
	sleep      time.Duration
	fossilized bool
}

func testFossilizeMultiple(t *testing.T, a *Fossilizer, tests []fossilizeTest) {
	go a.Start()
	defer a.Stop()
	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

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
	a := New(config)
	go a.Start()
	defer a.Stop()
	rc := make(chan *fossilizer.Result)
	a.AddResultChan(rc)

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
