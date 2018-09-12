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

// LinkSelector used in LinkQuery.
type LinkSelector struct {
	ObjectType   string        `json:"docType"`
	PrevLinkHash *PrevLinkHash `json:"linkWrapper.prevLinkHash,omitempty"`
	Process      string        `json:"linkWrapper.link.meta.process.name,omitempty"`
	MapIds       *MapIdsIn     `json:"linkWrapper.link.meta.mapId,omitempty"`
	Tags         *TagsAll      `json:"linkWrapper.link.meta.tags,omitempty"`
	LinkHash     *LinkHashIn   `json:"_id,omitempty"`
}

// LinkHashIn specifies the list of link hashes to search for
type LinkHashIn struct {
	LinkHashes []string `json:"$in,omitempty"`
}

// MapIdsIn specifies that segment mapId should be in specified list
type MapIdsIn struct {
	MapIds []string `json:"$in,omitempty"`
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
	Selector LinkSelector        `json:"selector,omitempty"`
	Limit    int                 `json:"limit,omitempty"`
	Skip     int                 `json:"skip,omitempty"`
	Sort     []map[string]string `json:"sort,omitempty"`
}

// CouchFindResponse is couchdb response type when posting to /db/_find
type CouchFindResponse struct {
	Docs []*Document `json:"docs"`
}

func buildSortArgs(reverse bool) []map[string]string {
	order := "desc"
	if reverse {
		order = "asc"
	}
	return []map[string]string{
		map[string]string{
			"linkWrapper.priority": order,
		},
	}
}

// NewSegmentQuery generates json data used to filter queries using couchdb _find api.
func NewSegmentQuery(filter *store.SegmentFilter) ([]byte, error) {
	linkSelector := LinkSelector{}
	linkSelector.ObjectType = objectTypeLink

	if filter.PrevLinkHash != nil {
		linkSelector.PrevLinkHash = &PrevLinkHash{
			Equals: *filter.PrevLinkHash,
		}
	}
	if filter.Process != "" {
		linkSelector.Process = filter.Process
	}
	if len(filter.MapIDs) > 0 {
		linkSelector.MapIds = &MapIdsIn{MapIds: filter.MapIDs}
	}
	if len(filter.Tags) > 0 {
		linkSelector.Tags = &TagsAll{Tags: filter.Tags}
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
		Sort:     buildSortArgs(filter.Reverse),
	}

	return json.Marshal(linkQuery)
}

// MapSelector used in MapQuery.
type MapSelector struct {
	ObjectType string         `json:"docType"`
	Process    string         `json:"process,omitempty"`
	MapIds     *MapIdsFilters `json:"_id,omitempty"`
}

// MapIdsFilters contain the filters on the segment map ID.
// MapIdsFilters.And is a list of MapIdsFilter.
type MapIdsFilters struct {
	Filters []MapIdsFilter `json:"$and,omitempty"`
}

// MapIdsFilter specifies that segment mapId should match a given regex.
type MapIdsFilter struct {
	MapIdsMatch string `json:"$regex,omitempty"`
}

// MapQuery used in CouchDB rich queries.
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

	mapIdsFilters := &MapIdsFilters{Filters: []MapIdsFilter{}}

	if filter.Prefix != "" {
		mapIdsFilters.Filters = append(
			mapIdsFilters.Filters,
			MapIdsFilter{MapIdsMatch: fmt.Sprintf("^%s", filter.Prefix)},
		)
	}
	if filter.Suffix != "" {
		mapIdsFilters.Filters = append(
			mapIdsFilters.Filters,
			MapIdsFilter{MapIdsMatch: fmt.Sprintf("%s$", filter.Suffix)},
		)
	}

	if len(mapIdsFilters.Filters) > 0 {
		mapSelector.MapIds = mapIdsFilters
	}

	mapQuery := MapQuery{
		Selector: mapSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	return json.Marshal(mapQuery)
}
