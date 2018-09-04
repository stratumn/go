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

package types

import (
	"sort"

	"github.com/stratumn/go-chainscript"
)

// SegmentSlice is a slice of segment pointers.
type SegmentSlice []*chainscript.Segment

// PaginatedSegments is a slice of segments along with the total results count.
type PaginatedSegments struct {
	Segments   SegmentSlice `json:"segments"`
	TotalCount int          `json:"totalCount"`
}

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
		p1 = s1.Link.Meta.Priority
		p2 = s2.Link.Meta.Priority
	)

	if p1 > p2 {
		return true
	}

	if p1 < p2 {
		return false
	}

	return s1.LinkHashString() < s2.LinkHashString()
}

// Sort returns the sorted segment slice.
func (s *SegmentSlice) Sort(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(s))
	} else {
		sort.Sort(s)
	}
}
