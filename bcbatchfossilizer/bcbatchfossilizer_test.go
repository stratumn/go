// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcbatchfossilizer

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/blockchain/dummytimestamper"
)

const interval = 10 * time.Millisecond

func TestGetInfo(t *testing.T) {
	a, err := New(&Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	got, err := a.GetInfo()
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	if _, ok := got.(*Info); !ok {
		t.Errorf("a.GetInfo(): info = %#v want *Info", got)
	}
}

func TestFossilize(t *testing.T) {
	a, err := New(&Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval: interval,
	})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	tests := []fossilizeTest{
		{atos(sha256.Sum256([]byte("a"))), []byte("test a"), pathABCDE0, 0, false},
		{atos(sha256.Sum256([]byte("b"))), []byte("test b"), pathABCDE1, 0, false},
		{atos(sha256.Sum256([]byte("c"))), []byte("test c"), pathABCDE2, 0, false},
		{atos(sha256.Sum256([]byte("d"))), []byte("test d"), pathABCDE3, 0, false},
		{atos(sha256.Sum256([]byte("e"))), []byte("test e"), pathABCDE4, 0, false},
	}
	testFossilizeMultiple(t, a, tests)
}
