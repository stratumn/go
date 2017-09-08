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

// Package cs defines types to work with Chainscripts.
package cs

import (
	"errors"
	"math"

	"reflect"

	"github.com/stratumn/sdk/types"
)

const (
	// CompleteEvidence characterizes a Complete Evidence.
	CompleteEvidence = "COMPLETE"

	// PendingEvidence characterizes a Pending Evidence.
	PendingEvidence = "PENDING"
)

// Segment contains a link and meta data about the link.
type Segment struct {
	Link Link                   `json:"link"`
	Meta map[string]interface{} `json:"meta"`
}

// GetLinkHash returns the link ID as bytes.
// It assumes the segment is valid.
func (s *Segment) GetLinkHash() *types.Bytes32 {
	b, _ := types.NewBytes32FromString(s.GetLinkHashString())
	return b
}

// GetLinkHashString returns the link ID as a string.
// It assumes the segment is valid.
func (s *Segment) GetLinkHashString() string {
	return s.Meta["linkHash"].(string)
}

// Validate checks for errors in a segment.
func (s *Segment) Validate() error {
	if linkHash, ok := s.Meta["linkHash"].(string); !ok || linkHash == "" {
		return errors.New("meta.linkHash should be a non empty string")
	}
	if process, ok := s.Link.Meta["process"].(string); !ok || process == "" {
		return errors.New("link.meta.process should be a non empty string")
	}
	if mapID, ok := s.Link.Meta["mapId"].(string); !ok || mapID == "" {
		return errors.New("link.meta.mapId should be a non empty string")
	}
	if v, ok := s.Link.Meta["prevLinkHash"]; ok {
		if prevLinkHash, ok := v.(string); !ok || prevLinkHash == "" {
			return errors.New("link.meta.prevLinkHash should be a non empty string")
		}
	}

	if v, ok := s.Link.Meta["tags"]; ok {
		tags, ok := v.([]interface{})
		if !ok {
			return errors.New("link.meta.tags should be an array of non empty string")
		}
		for _, t := range tags {
			if tag, ok := t.(string); !ok || tag == "" {
				return errors.New("link.meta.tags should be an array of non empty string")
			}
		}
	}

	if v, ok := s.Link.Meta["priority"]; ok {
		if _, ok := v.(float64); !ok {
			return errors.New("link.meta.priority should be a float64")
		}
	}

	return nil
}

// SetEvidence sets the segment evidence
func (s *Segment) SetEvidence(evidence map[string]interface{}) {
	s.Meta["evidence"] = evidence
}

// GetEvidence returns the segment evidence
func (s *Segment) GetEvidence() map[string]interface{} {
	return s.Meta["evidence"].(map[string]interface{})
}

// IsEmpty checks if a segment is empty (nil)
func (s *Segment) IsEmpty() bool {
	return reflect.DeepEqual(*s, Segment{})
}

// Link contains a state and meta data about the state.
type Link struct {
	State map[string]interface{} `json:"state"`
	Meta  map[string]interface{} `json:"meta"`
}

// GetPriority returns the priority as a float64
// It assumes the link is valid.
// If priority is nil, it will return -Infinity.
func (l *Link) GetPriority() float64 {
	if f64, ok := l.Meta["priority"].(float64); ok {
		return f64
	}
	return math.Inf(-1)
}

// GetMapID returns the map ID as a string.
// It assumes the link is valid.
func (l *Link) GetMapID() string {
	return l.Meta["mapId"].(string)
}

// GetPrevLinkHash returns the previous link hash as a bytes.
// It assumes the link is valid.
// It will return nilif the previous link hash is null.
func (l *Link) GetPrevLinkHash() *types.Bytes32 {
	if str, ok := l.Meta["prevLinkHash"].(string); ok {
		b, _ := types.NewBytes32FromString(str)
		return b
	}
	return nil
}

// GetPrevLinkHashString returns the previous link hash as a string.
// It assumes the link is valid.
// It will return an empty string if the previous link hash is null.
func (l *Link) GetPrevLinkHashString() string {
	if str, ok := l.Meta["prevLinkHash"].(string); ok {
		return str
	}
	return ""
}

// GetTags returns the tags as an array of string.
// It assumes the link is valid.
// It will return nil if there are no tags.
func (l *Link) GetTags() []string {
	if t, ok := l.Meta["tags"].([]interface{}); ok {
		tags := make([]string, len(t))
		for i, v := range t {
			tags[i] = v.(string)
		}
		return tags
	}
	return nil
}

// GetTagMap returns the tags as a map of string to empty structs (a set).
// It assumes the link is valid.
func (l *Link) GetTagMap() map[string]struct{} {
	tags := map[string]struct{}{}
	if t, ok := l.Meta["tags"].([]interface{}); ok {
		for _, v := range t {
			tags[v.(string)] = struct{}{}
		}
	}
	return tags
}

// GetProcess returns the process name as a string.
// It assumes the link is valid.
func (l *Link) GetProcess() string {
	if str, ok := l.Meta["process"].(string); ok {
		return str
	}
	return ""
}

// SegmentSlice is a slice of segment pointers.
type SegmentSlice []*Segment

// Len implements sort.Interface.Len.
func (s SegmentSlice) Len() int {
	return len(s)
}

// Swap implements sort.Interface.Swap.
func (s SegmentSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements sort.Interface.Less.
func (s SegmentSlice) Less(i, j int) bool {
	var (
		s1 = s[i]
		s2 = s[j]
		p1 = s1.Link.GetPriority()
		p2 = s2.Link.GetPriority()
	)

	if p1 > p2 {
		return true
	}

	if p1 < p2 {
		return false
	}

	return s1.GetLinkHashString() < s2.GetLinkHashString()
}
