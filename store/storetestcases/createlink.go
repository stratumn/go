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
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
)

// TestCreateLink tests what happens when you create a new link.
func (f Factory) TestCreateLink(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l := cstesting.RandomLink()
	if _, err := a.CreateLink(l); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}
}

// TestCreateLinkNoPriority tests what happens when you create a new link with no priority.
func (f Factory) TestCreateLinkNoPriority(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l := cstesting.RandomLink()
	delete(l.Meta, "priority")

	if _, err := a.CreateLink(l); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}
}

// TestCreateLinkUpdatedState tests what happens when you update the state of a
// link.
func (f Factory) TestCreateLinkUpdatedState(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l := cstesting.RandomLink()
	if _, err := a.CreateLink(l); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}

	l = cstesting.ChangeLinkState(l)
	if _, err := a.CreateLink(l); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}
}

// TestCreateLinkUpdatedMapID tests what happens when you update the map ID of
// a link.
func (f Factory) TestCreateLinkUpdatedMapID(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l1 := cstesting.RandomLink()
	if _, err := a.CreateLink(l1); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}

	l2 := cstesting.ChangeLinkMapID(l1)
	if _, err := a.CreateLink(l2); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}
}

// TestCreateLinkBranch tests what happens when you create a link with a
// previous link hash.
func (f Factory) TestCreateLinkBranch(t *testing.T) {
	a := f.initAdapterV2(t)
	defer f.freeV2(a)

	l := cstesting.RandomLink()
	if _, err := a.CreateLink(l); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}

	l = cstesting.RandomLinkBranch(l)
	if _, err := a.CreateLink(l); err != nil {
		t.Fatalf("a.CreateLink(): err: %s", err)
	}
}
