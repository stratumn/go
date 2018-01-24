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
	"fmt"
	"go/build"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Don't forget to set enableProcessUpload to true on the agent to test
// this feature.

// TestUploadProcessOK tests the client's ability to handle a CreateMap request.
func (f Factory) TestUploadProcessOK(t *testing.T) {
	process, err := f.Client.UploadProcess(
		"test",
		fmt.Sprintf("%v/src/github.com/stratumn/sdk/agent/agenttestcases/actions.js", build.Default.GOPATH),
		StoreURL,
		[]string{},
		[]string{},
	)

	assert.NoError(t, err)
	assert.NotNil(t, process)
}
