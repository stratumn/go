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

package agenttestcases

import (
	"testing"

	"github.com/stratumn/go-indigocore/types"
	"github.com/stretchr/testify/assert"
)

// TestGetSegmentOK tests the client's ability to handle a GetSegment request.
func (f Factory) TestGetSegmentOK(t *testing.T) {
	process := "test"
	parent, _ := f.Client.CreateMap(process, nil, "test")

	segment, err := f.Client.GetSegment(process, parent.GetLinkHash())
	assert.NoError(t, err)
	assert.NotNil(t, segment)
}

// TestGetSegmentNotFound tests the client's ability to handle a GetSegment request
// when the queried linkHash does not exist.
func (f Factory) TestGetSegmentNotFound(t *testing.T) {
	process := "test"
	fakeLinkHash, _ := types.NewBytes32FromString("0000000000000000000000000000000000000000000000000000000000000000")
	segment, err := f.Client.GetSegment(process, fakeLinkHash)
	assert.EqualError(t, err, "Not Found")
	assert.Nil(t, segment)
}
