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

func TestTMStore(t *testing.T) {
	storetestcases.Factory{
		New:  newTestTMStore,
		Free: freeTestTMStore,
	}.RunTests(t)
}

func newTestTMStore() (store.Adapter, error) {
	s := NewTestClient()
	s.RetryStartWebsocket(DefaultWsRetryInterval)

	return s, nil
}

func freeTestTMStore(s store.Adapter) {
	Reset()
}
