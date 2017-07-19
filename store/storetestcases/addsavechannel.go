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

package storetestcases

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
)

// TestAddDidSaveChannel tests that AddDidSaveChannel functions properly.
func (f Factory) TestAddDidSaveChannel(t *testing.T) {
	a := f.initAdapter(t)
	defer f.free(a)

	c := make(chan *cs.Segment, 1)
	a.AddDidSaveChannel(c)

	s := cstesting.RandomSegment()
	if err := a.SaveSegment(s); err != nil {
		t.Fatalf("a.SaveSegment(): err: %s", err)
	}

	if got, want := <-c, s; !reflect.DeepEqual(want, got) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(got, "", "  ")
		t.Errorf("<- c = %s\n want%s", gotJS, wantJS)
	}
}
