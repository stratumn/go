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

	t.Run("Subscribe", func(t *testing.T) {
		t.Run("Adds a listener provided with the current validator set", func(t *testing.T) {

			b := NewUpdateBroadcaster()
			b.Broadcast(validator)

			select {
			case <-b.Subscribe():
				break
			case <-time.After(10 * time.Millisecond):
				assert.Fail(t, "No validator in the channel")
			}
		})
	})
	t.Run("Unsubscribe", func(t *testing.T) {

		t.Run("Removes an unknown channel", func(t *testing.T) {
			b := NewUpdateBroadcaster()
			// try removing a non-existing channel. It should not panic.
			b.Unsubscribe(make(chan validators.Validator))

			// add a channel and retry removing a non-existing channel. It should not panic.
			b.Subscribe()
			b.Unsubscribe(make(chan validators.Validator))
		})

		t.Run("Closes the channel", func(t *testing.T) {
			b := NewUpdateBroadcaster()
			listener := b.Subscribe()
			b.Unsubscribe(listener)

			_, ok := <-listener
			assert.False(t, ok, "<-listener")
		})

	})

	t.Run("Broadcast", func(t *testing.T) {
		b := NewUpdateBroadcaster()
		listener1 := b.Subscribe()
		listener2 := b.Subscribe()

		b.Broadcast(validator)

		for i := 0; i < 2; i++ {
			select {
			case <-listener1:
				continue
			case <-listener2:
				continue
			case <-time.After(10 * time.Millisecond):
				assert.Fail(t, "No validator in the channel")
			}
		}
	})

	t.Run("Close", func(t *testing.T) {
		b := NewUpdateBroadcaster()
		listener1 := b.Subscribe()
		listener2 := b.Subscribe()

		b.Close()

		for i := 0; i < 2; i++ {
			select {
			case _, ok := <-listener1:
				assert.False(t, ok, "<-listener")
			case _, ok := <-listener2:
				assert.False(t, ok, "<-listener")
			case <-time.After(10 * time.Millisecond):
				assert.Fail(t, "No validator in the channel")
			}
		}
	})
}
