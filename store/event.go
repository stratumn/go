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

// Package store defines types to implement a store.
package store

import (
	"encoding/hex"
	"encoding/json"

	"github.com/stratumn/go-chainscript"
)

// EventType lets you know the kind of event received.
// A client should ignore events it doesn't care about or doesn't understand.
type EventType string

const (
	// SavedLinks means that segment links were saved.
	SavedLinks EventType = "SavedLinks"
	// SavedEvidences means that segment evidences were saved.
	SavedEvidences EventType = "SavedEvidences"
)

// Event is the object stores send to notify of important events.
type Event struct {
	EventType EventType
	Data      interface{}
}

// NewSavedLinks creates a new event to notify links were saved.
func NewSavedLinks(linkArgs ...*chainscript.Link) *Event {
	links := make([]*chainscript.Link, 0, len(linkArgs))
	links = append(links, linkArgs...)
	return &Event{
		EventType: SavedLinks,
		Data:      links,
	}
}

// AddSavedLinks adds links to the event.
// It assumes the event is a correctly initialized SavedLinks event.
func (event *Event) AddSavedLinks(links ...*chainscript.Link) {
	linksData := event.Data.([]*chainscript.Link)
	linksData = append(linksData, links...)
	event.Data = linksData
}

// NewSavedEvidences creates a new event to notify evidences were saved.
func NewSavedEvidences() *Event {
	evidences := make(map[string]*chainscript.Evidence)
	return &Event{
		EventType: SavedEvidences,
		Data:      evidences,
	}
}

// AddSavedEvidence adds an evidence to the event.
// It assumes the event is a correctly initialized SavedEvidences event.
func (event *Event) AddSavedEvidence(linkHash []byte, e *chainscript.Evidence) {
	evidencesData := event.Data.(map[string]*chainscript.Evidence)
	evidencesData[hex.EncodeToString(linkHash)] = e
	event.Data = evidencesData
}

// UnmarshalJSON does custom deserialization to correctly type the Data field.
func (event *Event) UnmarshalJSON(b []byte) error {
	partial := struct {
		EventType EventType
		Data      json.RawMessage
	}{}

	if err := json.Unmarshal(b, &partial); err != nil {
		return err
	}

	var data interface{}
	switch partial.EventType {
	case SavedLinks:
		var links []*chainscript.Link
		if err := json.Unmarshal(partial.Data, &links); err != nil {
			return err
		}
		data = links
	case SavedEvidences:
		var evidences map[string]*chainscript.Evidence
		if err := json.Unmarshal(partial.Data, &evidences); err != nil {
			return err
		}
		data = evidences
	}

	*event = Event{
		EventType: partial.EventType,
		Data:      data,
	}

	return nil
}
