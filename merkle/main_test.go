// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"flag"
	"os"
	"testing"

	"github.com/stratumn/goprivate/merkle/treetestcases"
)

func TestMain(m *testing.M) {
	treetestcases.LoadFixtures("testdata")
	flag.Parse()
	os.Exit(m.Run())
}
