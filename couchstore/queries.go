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

package couchstore

import (
	"encoding/json"
	"fmt"

	"github.com/stratumn/go-indigocore/store"
)

// LinkSelector used in LinkQuery
type LinkSelector struct {
	ObjectType   string         `json:"docType"`
	PrevLinkHash *PrevLinkHash  `json:"link.meta.prevLinkHash,omitempty"`
	Process      string         `json:"link.meta.process,omitempty"`
	MapIds       *MapIdsFilters `json:"link.meta.mapId,omitempty"`
	Tags         *TagsAll       `json:"link.meta.tags,omitempty"`
	LinkHash     *LinkHashIn    `json:"_id,omitempty"`
}

// LinkHashIn specifies the list of link hashes to search for
type LinkHashIn struct {
	LinkHashes []string `json:"$in,omitempty"`
}

// MapIdsFilters contain the filters on the segment map ID.
// MapIdsFilters.And is a list of MapIdsIn and MapIdsMatch.
type MapIdsFilters struct {
	Filters []MapIdsFilter `json:"$and,omitempty"`
}

// MapIdsFilter specifies that segment mapId should be in specified list
// and should match a given regex.
type MapIdsFilter struct {
	MapIdsIn    []string `json:"$in,omitempty"`
	MapIdsMatch string   `json:"$regex,omitempty"`
}

// TagsAll specifies all tags in specified list should be in segment tags
type TagsAll struct {
	Tags []string `json:"$all,omitempty"`
}

// PrevLinkHash is used to specify PrevLinkHash in selector.
type PrevLinkHash struct {
	Exists *bool  `json:"$exists,omitempty"`
	Equals string `json:"$eq"`
}

// LinkQuery used in CouchDB rich queries
type LinkQuery struct {
	Selector LinkSelector `json:"selector,omitempty"`
	Limit    int          `json:"limit,omitempty"`
	Skip     int          `json:"skip,omitempty"`
}

// CouchFindResponse is couchdb response type when posting to /db/_find
type CouchFindResponse struct {
	Docs []*Document `json:"docs"`
}

// NewSegmentQuery generates json data used to filter queries using couchdb _find api.
func NewSegmentQuery(filter *store.SegmentFilter) ([]byte, error) {
	linkSelector := LinkSelector{}
	linkSelector.ObjectType = objectTypeLink

	mapIdsFilters := &MapIdsFilters{Filters: []MapIdsFilter{}}
	if len(filter.MapIDs) > 0 {
		mapIdsFilters.Filters = append(
			mapIdsFilters.Filters,
			MapIdsFilter{MapIdsIn: filter.MapIDs},
		)
	}
	if filter.MapIDPrefix != "" {
		mapIdsFilters.Filters = append(
			mapIdsFilters.Filters,
			MapIdsFilter{MapIdsMatch: fmt.Sprintf("^%s", filter.MapIDPrefix)},
		)
	}
	if filter.MapIDSuffix != "" {
		mapIdsFilters.Filters = append(
			mapIdsFilters.Filters,
			MapIdsFilter{MapIdsMatch: fmt.Sprintf("%s$", filter.MapIDSuffix)},
		)
	}
	if len(mapIdsFilters.Filters) > 0 {
		linkSelector.MapIds = mapIdsFilters
	}

	if filter.PrevLinkHash != nil {
		linkSelector.PrevLinkHash = &PrevLinkHash{
			Equals: *filter.PrevLinkHash,
		}
	}
	if filter.Process != "" {
		linkSelector.Process = filter.Process
	}
	if len(filter.Tags) > 0 {
		linkSelector.Tags = &TagsAll{Tags: filter.Tags}
	} else {
		linkSelector.Tags = nil
	}
	if len(filter.LinkHashes) > 0 {
		linkSelector.LinkHash = &LinkHashIn{
			LinkHashes: filter.LinkHashes,
		}
	}

	linkQuery := LinkQuery{
		Selector: linkSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	return json.Marshal(linkQuery)
}

// MapSelector used in MapQuery
type MapSelector struct {
	ObjectType string `json:"docType"`
	Process    string `json:"process,omitempty"`
}

// MapQuery used in CouchDB rich queries
type MapQuery struct {
	Selector MapSelector `json:"selector,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Skip     int         `json:"skip,omitempty"`
}

// NewMapQuery generates json data used to filter queries using couchdb _find api.
func NewMapQuery(filter *store.MapFilter) ([]byte, error) {
	mapSelector := MapSelector{}
	mapSelector.ObjectType = objectTypeMap
	mapSelector.Process = filter.Process

	mapQuery := MapQuery{
		Selector: mapSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	return json.Marshal(mapQuery)
}
