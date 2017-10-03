// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package bcbatchfossilizer

import (
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/blockchain/dummytimestamper"
	"github.com/stratumn/sdk/cs/evidences"
)

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
func TestBcBatchProof(t *testing.T) {
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
	results := testFossilizeMultiple(t, a, tests)

	t.Run("TestTime()", func(t *testing.T) {
		for _, r := range results {
			e := r.Evidence.Proof.(*evidences.BcBatchProof)
			if e.Time() != uint64(time.Now().Unix()) {
				t.Errorf("wrong timestamp in BcBatchProof")
			}
		}
	})

	t.Run("TestFullProof()", func(t *testing.T) {
		for _, r := range results {
			e := r.Evidence.Proof.(*evidences.BcBatchProof)
			p := e.FullProof()
			if p == nil {
				t.Errorf("got evidence.FullProof() == nil")
			}
			if err := json.Unmarshal(p, &evidences.BcBatchProof{}); err != nil {
				t.Errorf("Could not unmarshal bytes proof, err = %+v", err)
			}
		}
	})

	t.Run("TestVerify()", func(t *testing.T) {
		for _, r := range results {
			e := r.Evidence.Proof.(*evidences.BcBatchProof)
			if e.Verify(nil) != true {
				t.Errorf("got evidence.Verify() == false")
			}
		}
	})
}
