// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmstore

import (
	"testing"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/store/storetestcases"
)

var (
	tmstore *TMStore
	test    *testing.T
)

func TestTMStore(t *testing.T) {
	test = t
	storetestcases.Factory{
		New:  newTestTMStore,
		Free: freeTestTMStore,
	}.RunTests(t)
}

func newTestTMStore() (store.Adapter, error) {
	tmstore = NewTestClient()
	tmstore.RetryStartWebsocket(DefaultWsRetryInterval)

	return tmstore, nil
}

func freeTestTMStore(s store.Adapter) {
	mapIDs, err := tmstore.GetMapIDs(&store.Pagination{Limit: 100})
	if err != nil {
		test.Fatal(err)
	}
	for _, m := range mapIDs {
		segments, err := tmstore.FindSegments(&store.Filter{MapID: m, Pagination: store.Pagination{Limit: 100}})
		if err != nil {
			test.Fatal(err)
		}
		for _, s := range segments {
			tmstore.DeleteSegment(s.GetLinkHash())
		}
	}
}
