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
	"sync"

	"github.com/stratumn/go-indigocore/validation/validators"
)

// UpdateSubscriber provides a way to be notified of validation updates.
type UpdateSubscriber interface {
	Subscribe() <-chan validators.Validator
	Unsubscribe(<-chan validators.Validator)
}

// UpdateNotifier allows broadcasting a validator to a bunch of subscribers.
type UpdateNotifier interface {
	Broadcast(validators.Validator)
	Close()
}

// UpdateBroadcaster implements github.com/go-indigocore/validation.UpdateSubscriber and github.com/go-indigocore/validation.UpdateNotifier.
// It provides subscription to a Manager to be notified of validation updates.
type UpdateBroadcaster struct {
	current validators.Validator

	listenersMutex sync.RWMutex
	listeners      []chan validators.Validator
}

// NewUpdateBroadcaster returns a new UpdateBroadcaster.
func NewUpdateBroadcaster() *UpdateBroadcaster {
	return &UpdateBroadcaster{}
}

// Subscribe implements github.com/go-indigocore/validation.UpdateSubscriber.Subscribe.
// It return a listener that will be notified when the validator changes.
func (b *UpdateBroadcaster) Subscribe() <-chan validators.Validator {
	b.listenersMutex.Lock()
	defer b.listenersMutex.Unlock()

	subscribeChan := make(chan validators.Validator)
	b.listeners = append(b.listeners, subscribeChan)
	// Insert the current validator in the channel if there is one.
	if b.current != nil {
		go func() {
			subscribeChan <- b.current
		}()
	}
	return subscribeChan
}

// Unsubscribe implements github.com/go-indigocore/validation.UpdateSubscriber.Unsubscribe.
// It removes a listener.
func (b *UpdateBroadcaster) Unsubscribe(c <-chan validators.Validator) {
	b.listenersMutex.Lock()
	defer b.listenersMutex.Unlock()

	index := -1
	for i, l := range b.listeners {
		if l == c {
			index = i
			break
		}
	}

	if index >= 0 {
		close(b.listeners[index])
		b.listeners[index] = b.listeners[len(b.listeners)-1]
		b.listeners = b.listeners[:len(b.listeners)-1]
	}
}

// Broadcast implements github.com/go-indigocore/validation.UpdateNotifier.Broadcast.
func (b *UpdateBroadcaster) Broadcast(validator validators.Validator) {
	b.listenersMutex.RLock()
	defer b.listenersMutex.RUnlock()

	b.current = validator
	for _, listener := range b.listeners {
		go func(listener chan validators.Validator) {
			listener <- validator
		}(listener)
	}
}

// Close implements github.com/go-indigocore/validation.UpdateNotifier.Close.
func (b *UpdateBroadcaster) Close() {
	b.listenersMutex.Lock()
	defer b.listenersMutex.Unlock()

	for _, s := range b.listeners {
		close(s)
	}
}
