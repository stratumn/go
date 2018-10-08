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

// Package fossilizer defines types to implement a fossilizer.
package fossilizer

import (
	"context"

	"github.com/stratumn/go-chainscript"
)

// Adapter must be implemented by a fossilier.
type Adapter interface {
	// Returns arbitrary information about the adapter.
	GetInfo(context.Context) (interface{}, error)

	// Adds a channel that receives events from the fossilizer
	AddFossilizerEventChan(chan *Event)

	// Requests data to be fossilized.
	// Meta is arbitrary data that will be forwarded to the websocket.
	Fossilize(ctx context.Context, data []byte, meta []byte) error
}

// Result is the type sent to the result channels.
type Result struct {
	// Evidence created by the fossilizer.
	Evidence chainscript.Evidence

	// The data that was fossilized.
	Data []byte

	// The meta data that was given to Adapter.Fossilize.
	Meta []byte
}

// EventType lets you know the kind of event received.
// A client should ignore events it doesn't care about or doesn't understand.
type EventType string

const (
	// DidFossilizeLink means that the link was fossilized
	DidFossilizeLink EventType = "DidFossilizeLink"
)

// Event is the object fossilizers send to notify of important events.
type Event struct {
	EventType EventType
	Data      interface{}
}
