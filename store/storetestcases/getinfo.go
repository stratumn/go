// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

package storetestcases

import (
	"context"
	"testing"
)

// TestGetInfo tests what happens when you get information about the adapter.
func (f Factory) TestGetInfo(t *testing.T) {
	a := f.initAdapter(t)
	defer f.freeAdapter(a)

	info, err := a.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("a.GetInfo(): err: %s", err)
	}
	if info == nil {
		t.Fatal("info = nil want interface{}")
	}
}
