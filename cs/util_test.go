// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cs_test

import (
	"testing"

	"github.com/stratumn/sdk/cs"
)

func testSegmentValidateError(t *testing.T, s *cs.Segment, want string) {
	if err := s.Validate(); err == nil {
		t.Error("s.Validate() = nil want Error")
	} else {
		if got := err.Error(); got != want {
			t.Errorf("s.Validate() = %q want %q", got, want)
		}
	}
}
