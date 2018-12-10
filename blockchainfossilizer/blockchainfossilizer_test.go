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

package blockchainfossilizer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stratumn/go-core/blockchainfossilizer/evidences"

	"github.com/stratumn/go-core/blockchain/blockchaintesting"
	"github.com/stratumn/go-core/blockchain/btc/btctimestamper"
	"github.com/stratumn/go-core/blockchainfossilizer"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	ts, err := btctimestamper.New(&btctimestamper.Config{
		WIF: "cMptgcyVp9nPpmvWM9tSR6SyCMkGH4xUX1LkJ2ZTTwfUfbZGXfXB",
	})
	require.NoError(t, err)

	a := blockchainfossilizer.New(&blockchainfossilizer.Config{
		Commit:      "commit",
		Version:     "version",
		Timestamper: ts,
	})

	resp, err := a.GetInfo(context.Background())
	require.NoError(t, err)

	info, ok := resp.(*blockchainfossilizer.Info)
	require.True(t, ok)
	assert.Equal(t, "commit", info.Commit)
	assert.Equal(t, "version", info.Version)
	assert.Equal(t, blockchainfossilizer.Name, info.Name)
	assert.Equal(t, "Stratumn's Blockchain Fossilizer with Bitcoin Timestamper", info.Description)
	assert.Equal(t, "bitcoin:test3", info.Blockchain)
}

func TestFossilize(t *testing.T) {
	t.Run("timestamper error", func(t *testing.T) {
		ts := blockchaintesting.NewTimestamper().WithError(errors.New("fatal, faut que j'te parle"))
		a := blockchainfossilizer.New(&blockchainfossilizer.Config{
			Timestamper: ts,
		})

		err := a.Fossilize(context.Background(), []byte{42}, []byte{24})
		assert.EqualError(t, err, "fatal, faut que j'te parle")
	})

	t.Run("timestamper success", func(t *testing.T) {
		ts := blockchaintesting.NewTimestamper().WithTransactionID([]byte{51})
		a := blockchainfossilizer.New(&blockchainfossilizer.Config{
			Timestamper: ts,
		})

		eventChan := make(chan *fossilizer.Event)
		a.AddFossilizerEventChan(eventChan)

		err := a.Fossilize(context.Background(), []byte{42}, []byte{24})
		require.NoError(t, err)

		e := <-eventChan
		assert.Equal(t, fossilizer.DidFossilize, e.EventType)

		r := e.Data.(*fossilizer.Result)
		assert.Equal(t, []byte{42}, r.Data)
		assert.Equal(t, []byte{24}, r.Meta)
		assert.Equal(t, blockchainfossilizer.Name, r.Evidence.Backend)
		assert.Equal(t, ts.GetInfo().Network.String(), r.Evidence.Provider)

		var bcProof evidences.BlockchainProof
		require.NoError(t, json.Unmarshal(r.Evidence.Proof, &bcProof))
		assert.Equal(t, []byte{42}, bcProof.Data)
		assert.Equal(t, []byte{51}, []byte(bcProof.TransactionID))
		assert.InDelta(t, time.Now().Unix(), bcProof.Timestamp, 10)
	})
}
