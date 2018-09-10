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
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/batchfossilizer"
	"github.com/stratumn/go-indigocore/bcbatchfossilizer/evidences"
	"github.com/stratumn/go-indigocore/blockchain/dummytimestamper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	a, err := New(&Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{})
	if err != nil {
		t.Fatalf("New(): err: %s", err)
	}
	got, err := a.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	info, ok := got.(*Info)
	if !ok {
		t.Errorf("a.GetInfo(): info = %#v want *Info", got)
	}
	if got, want := info.Description, "Indigo's Blockchain Batch Fossilizer with Dummy Timestamper"; got != want {
		t.Errorf("a.GetInfo(): Description = %s want %s", got, want)
	}
}

func TestFossilize(t *testing.T) {
	a, err := New(&Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval: testInterval,
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
		Interval: testInterval,
	})
	require.NoError(t, err)

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
			p, err := evidences.UnmarshalProof(&r.Evidence)
			require.NoError(t, err)

			assert.Equal(t, uint64(time.Now().Unix()), p.Time())
		}
	})

	t.Run("TestVerify()", func(t *testing.T) {
		for _, r := range results {
			p, err := evidences.UnmarshalProof(&r.Evidence)
			require.NoError(t, err)

			assert.True(t, p.Verify(nil))
		}
	})
}
