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

// Package cstesting defines helpers to test Chainscripts.
package cstesting

import (
	"encoding/json"
	"math/rand"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/testutil"
)

// CreateSegment creates a minimal segment.
func CreateSegment(process, linkHash, mapID, prevLinkHash string, tags []interface{}, priority float64) *cs.Segment {
	linkMeta := map[string]interface{}{
		"process":  process,
		"mapId":    mapID,
		"priority": priority,
		"random":   testutil.RandomString(12),
	}

	if prevLinkHash != "" {
		linkMeta["prevLinkHash"] = prevLinkHash
	}

	if tags != nil {
		linkMeta["tags"] = tags
	}

	return &cs.Segment{
		Link: cs.Link{
			State: map[string]interface{}{
				"random": testutil.RandomString(12),
			},
			Meta: linkMeta,
		},
		Meta: map[string]interface{}{
			"linkHash": linkHash,
			"random":   testutil.RandomString(12),
		},
	}
}

// RandomSegment creates a random segment.
func RandomSegment() *cs.Segment {
	return CreateSegment(testutil.RandomString(24), testutil.RandomHash().String(), testutil.RandomString(24),
		testutil.RandomHash().String(), RandomTags(), rand.Float64())
}

// ChangeSegmentState clones a segment and randomly changes its state.
func ChangeSegmentState(s *cs.Segment) *cs.Segment {
	clone := CloneSegment(s)
	clone.Link.State["random"] = testutil.RandomString(12)
	return clone
}

// ChangeSegmentMapID clones a segment and randomly changes its map ID.
func ChangeSegmentMapID(s *cs.Segment) *cs.Segment {
	clone := CloneSegment(s)
	clone.Link.Meta["mapId"] = testutil.RandomString(24)
	return clone
}

// RandomBranch appends a random segment to a segment.
func RandomBranch(s *cs.Segment) *cs.Segment {
	branch := CreateSegment(testutil.RandomString(24), testutil.RandomHash().String(), testutil.RandomString(24),
		s.Meta["linkHash"].(string), RandomTags(), rand.Float64())
	branch.Link.Meta["mapId"] = s.Link.Meta["mapId"]
	return branch
}

// RandomTags creates between zero and four random tags.
func RandomTags() []interface{} {
	var tags []interface{}
	for i := 0; i < rand.Intn(5); i++ {
		tags = append(tags, testutil.RandomString(12))
	}
	return tags
}

// CloneSegment clones a segment.
func CloneSegment(s *cs.Segment) *cs.Segment {
	var clone cs.Segment

	js, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(js, &clone); err != nil {
		panic(err)
	}

	return &clone
}
