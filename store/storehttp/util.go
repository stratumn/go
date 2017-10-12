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

package storehttp

import (
	"net/http"
	"strconv"

	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

func parseSegmentFilter(r *http.Request) (*store.SegmentFilter, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return nil, err
	}

	const prevLinkHashKey = "prevLinkHash"

	var (
		q               = r.URL.Query()
		mapIDs          = append(q["mapIds[]"], q["mapIds%5B%5D"]...)
		process         = q.Get("process")
		prevLinkHashStr = q.Get(prevLinkHashKey)
		tags            = append(q["tags[]"], q["tags%5B%5D"]...)
		prevLinkHash    *string
	)

	if _, exists := q[prevLinkHashKey]; exists {
		prevLinkHash = &prevLinkHashStr
		if *prevLinkHash != "" {
			_, err := types.NewBytes32FromString(*prevLinkHash)
			if err != nil {
				return nil, newErrPrevLinkHash("")
			}
		}
	}

	return &store.SegmentFilter{
		Pagination:   *pagination,
		MapIDs:       mapIDs,
		Process:      process,
		PrevLinkHash: prevLinkHash,
		Tags:         tags,
	}, nil
}

func parseMapFilter(r *http.Request) (*store.MapFilter, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return nil, err
	}

	var process = r.URL.Query().Get("process")

	return &store.MapFilter{
		Pagination: *pagination,
		Process:    process,
	}, nil
}

func parsePagination(r *http.Request) (*store.Pagination, error) {
	var err error

	q := r.URL.Query()
	offsetstr := q.Get("offset")
	offset := 0
	if offsetstr != "" {
		if offset, err = strconv.Atoi(offsetstr); err != nil || offset < 0 {
			return nil, newErrOffset("")
		}
	}

	limitstr := q.Get("limit")
	limit := store.DefaultLimit
	if limitstr != "" {
		if limit, err = strconv.Atoi(limitstr); err != nil || limit < 0 {
			return nil, newErrLimit("")
		}
	}

	if limit > store.MaxLimit {
		return nil, newErrLimit("")
	}

	return &store.Pagination{
		Offset: offset,
		Limit:  limit,
	}, nil
}
