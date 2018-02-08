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

	"github.com/stratumn/go-indigocore/agent"
	"github.com/stretchr/testify/assert"
)

// TestGetInfoOK tests the client's ability to handle a GetInfo request.
func (f Factory) TestGetInfoOK(t *testing.T) {
	actual, err := f.Client.GetInfo()

	expected := agent.Info{
		Processes: agent.ProcessesMap{
			"test": &agent.Process{},
		},
		Stores: []agent.StoreInfo{
			agent.StoreInfo{
				"url": StoreURL,
			},
		},
		Fossilizers: []agent.FossilizerInfo{},
		Plugins:     []agent.PluginInfo{},
	}
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expected.Stores, actual.Stores)
	assert.Equal(t, expected.Plugins, actual.Plugins)
	assert.Equal(t, expected.Fossilizers, actual.Fossilizers)
	assert.NotNil(t, expected.Processes["test"])
	assert.Equal(t, len(expected.Processes), len(actual.Processes))
}
