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

package validation

import (
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/validation/validators"
	"github.com/stretchr/testify/assert"
)

func TestBroadcaster(t *testing.T) {

	validator := validators.NewMultiValidator(nil)

	t.Run("AddRemoveListener", func(t *testing.T) {
		t.Run("Adds a listener provided with the current valitor set", func(t *testing.T) {

			b := Broadcaster{}
			b.broadcast(validator)

			select {
			case <-b.AddListener():
				break
			case <-time.After(10 * time.Millisecond):
				t.Error("No validator in the channel")
			}
		})

		t.Run("Removes an unknown channel", func(t *testing.T) {
			b := Broadcaster{}
			b.RemoveListener(make(chan validators.Validator))
			b.AddListener()
			b.RemoveListener(make(chan validators.Validator))
		})

		t.Run("Removes closes the channel", func(t *testing.T) {
			b := Broadcaster{}
			listener := b.AddListener()
			b.RemoveListener(listener)

			_, ok := <-listener
			assert.False(t, ok, "<-listener")
		})
	})

}
