package couchstore

import (
	"encoding/json"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

// SegmentSelector used in SegmentQuery
type SegmentSelector struct {
	ObjectType   string        `json:"docType"`
	PrevLinkHash *PrevLinkHash `json:"segment.link.meta.prevLinkHash,omitempty"`
	Process      string        `json:"segment.link.meta.process,omitempty"`
	MapIds       *MapIdsIn     `json:"segment.link.meta.mapId,omitempty"`
	Tags         *TagsAll      `json:"segment.link.meta.tags,omitempty"`
	LinkHash     *LinkHashIn   `json:"_id,omitempty"`
}

// LinkHashIn specifies the list of link hashes to search for
type LinkHashIn struct {
	LinkHashes []*types.Bytes32 `json:"$in,omitempty"`
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
	Equals string `json:"$eq,omitempty"`
}

// SegmentQuery used in CouchDB rich queries
type SegmentQuery struct {
	Selector SegmentSelector `json:"selector,omitempty"`
	Limit    int             `json:"limit,omitempty"`
	Skip     int             `json:"skip,omitempty"`
}

// CouchFindResponse is couchdb response type when posting to /db/_find
type CouchFindResponse struct {
	Docs []*Document `json:"docs"`
}

// NewSegmentQuery generates json data used to filter queries using couchdb _find api.
func NewSegmentQuery(filter *store.SegmentFilter) ([]byte, error) {
	segmentSelector := SegmentSelector{}
	segmentSelector.ObjectType = objectTypeSegment

	if filter.PrevLinkHash != nil {
		if *filter.PrevLinkHash == "" {
			no := false
			segmentSelector.PrevLinkHash = &PrevLinkHash{
				Exists: &no,
			}
		} else {
			segmentSelector.PrevLinkHash = &PrevLinkHash{
				Equals: *filter.PrevLinkHash,
			}
		}
	}
	if filter.Process != "" {
		segmentSelector.Process = filter.Process
	}
	if len(filter.MapIDs) > 0 {
		segmentSelector.MapIds = &MapIdsIn{filter.MapIDs}
	} else {
		segmentSelector.MapIds = nil
	}
	if len(filter.Tags) > 0 {
		segmentSelector.Tags = &TagsAll{filter.Tags}
	} else {
		segmentSelector.Tags = nil
	}
	if len(filter.LinkHashes) > 0 {
		segmentSelector.LinkHash = &LinkHashIn{
			LinkHashes: filter.LinkHashes,
		}
	}

	segmentQuery := SegmentQuery{
		Selector: segmentSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	return json.Marshal(segmentQuery)
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

	if filter.Process != "" {
		mapSelector.Process = filter.Process
	}

	mapQuery := MapQuery{
		Selector: mapSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	return json.Marshal(mapQuery)
}
