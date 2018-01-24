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

	"github.com/stretchr/testify/assert"
)

// TestGetProcessOK tests the client's ability to handle a GetProcess request.
func (f Factory) TestGetProcessOK(t *testing.T) {
	process, err := f.Client.GetProcess("test")
	assert.NoError(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, "test", process.Name)
	assert.Equal(t, 2, len(process.ProcessInfo.Actions))
}

// TestGetProcessNotFound tests the client's ability to handle a FindSegment request
// when no process is found.
func (f Factory) TestGetProcessNotFound(t *testing.T) {
	process, err := f.Client.GetProcess("wrong")
	assert.EqualError(t, err, "Not Found")
	assert.Nil(t, process)
}
