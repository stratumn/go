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

package tmpoptestcases

import (
	"bytes"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/tmpop/tmpoptestcases/mocks"
	"github.com/stretchr/testify/assert"
)

// TestCheckTx tests what happens when the ABCI method CheckTx() is called
func (f Factory) TestCheckTx(t *testing.T) {
	h, _ := f.newTMPop(t, nil)
	defer f.free()

	t.Run("Check valid link returns ok", func(t *testing.T) {
		_, tx := makeCreateRandomLinkTx(t)
		res := h.CheckTx(tx)

		if !res.IsOK() {
			t.Errorf("Expected CheckTx to return an OK result, got %v", res)
		}
	})

	t.Run("Check link with invalid reference returns not-ok", func(t *testing.T) {
		link := cstesting.RandomLink()
		link.Meta["refs"] = []interface{}{map[string]interface{}{
			"process":  "proc",
			"linkHash": "invalidLinkHash",
		}}
		tx := makeCreateLinkTx(t, link)

		res := h.CheckTx(tx)

		if res.Code != tmpop.CodeTypeValidation {
			t.Errorf("Expected CheckTx to return an error, got %v", res)
		}
	})

	t.Run("Check link with uncommitted but checked reference returns ok", func(t *testing.T) {
		link, tx := makeCreateRandomLinkTx(t)
		linkHash, _ := link.Hash()
		res := h.CheckTx(tx)

		linkWithRef := cstesting.RandomLinkWithProcess(link.GetProcess())
		linkWithRef.Meta["refs"] = []interface{}{map[string]interface{}{
			"process":  link.GetProcess(),
			"linkHash": linkHash,
		}}
		tx = makeCreateLinkTx(t, linkWithRef)

		res = h.CheckTx(tx)

		if !res.IsOK() {
			t.Errorf("Expected CheckTx to return an OK result, got %v", res)
		}
	})
}

// TestDeliverTx tests what happens when the ABCI method DeliverTx() is called
func (f Factory) TestDeliverTx(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	h.BeginBlock(req)

	t.Run("Deliver valid link returns ok", func(t *testing.T) {
		_, tx := makeCreateRandomLinkTx(t)
		res := h.DeliverTx(tx)

		if !res.IsOK() {
			t.Errorf("Expected DeliverTx to return an OK result, got %v", res)
		}
	})

	t.Run("Deliver link referencing a checked but not delivered link returns an error", func(t *testing.T) {
		link, tx := makeCreateRandomLinkTx(t)
		linkHash, _ := link.Hash()
		res := h.CheckTx(tx)

		linkWithRef := cstesting.RandomLinkWithProcess(link.GetProcess())
		linkWithRef.Meta["refs"] = []interface{}{map[string]interface{}{
			"process":  link.GetProcess(),
			"linkHash": linkHash,
		}}
		tx = makeCreateLinkTx(t, linkWithRef)
		res = h.DeliverTx(tx)

		if res.Code != tmpop.CodeTypeValidation {
			t.Errorf("Expected DeliverTx to return an error, got %v", res)
		}
	})
}

// TestCommitTx tests what happens when the ABCI method CommitTx() is called
func (f Factory) TestCommitTx(t *testing.T) {
	h, req := f.newTMPop(t, nil)
	defer f.free()

	tmClientMock := new(tmpoptestcasesmocks.MockedTendermintClient)
	tmClientMock.AllowCalls()

	h.ConnectTendermint(tmClientMock)

	previousAppHash := req.Header.AppHash
	h.BeginBlock(req)

	link1, tx := makeCreateRandomLinkTx(t)
	h.DeliverTx(tx)

	link2, tx := makeCreateRandomLinkTx(t)
	h.DeliverTx(tx)

	res := h.Commit()
	if !res.IsOK() {
		t.Fatalf("Commit failed: %v", res)
	}

	t.Run("Commit correctly saves links and updates app hash", func(t *testing.T) {
		verifyLinkStored(t, h, link1)
		verifyLinkStored(t, h, link2)

		if bytes.Compare(previousAppHash, res.Data) == 0 {
			t.Errorf("Committed app hash is the same as the previous app hash")
		}
	})

	t.Run("Commit fires an event to notify that links were created", func(t *testing.T) {
		tmClientMock.AssertNumberOfCalls(t, "FireEvent", 1)
		eventsCall := tmClientMock.Calls[0]
		events := eventsCall.Arguments.Get(1).(tmpop.StoreEventsData)
		assert.Exactly(t, 2, len(events.StoreEvents), "Invalid number of events")
		assert.EqualValues(t, link1, events.StoreEvents[0].Details, "Invalid event details")
		assert.EqualValues(t, link2, events.StoreEvents[1].Details, "Invalid event details")
	})
}
