// Copyright 2016 Stratumn SAS. All rights reserved.
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

// Package storehttp is used to create an HTTP server from a store adapter.
//
// It serves the following routes:
//	GET /
//		Renders information about the fossilizer.
//
//	POST /segments
//		Saves then renders a segment.
//		Body should be a JSON encoded segment.
//
//	GET /segments/:linkHash
//		Renders a segment.
//
//	DELETE /segments/:linkHash
//		Deletes then renders a segment.
//
//	GET /segments?[offset=offset]&[limit=limit]&[mapId=mapId]&[prevLinkHash=prevLinkHash]&[tags=list+of+tags]
//		Finds and renders segments.
//
//	GET /maps?[offset=offset]&[limit=limit]
//		Finds and renders map IDs.
package storehttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/types"
)

const (
	// DefaultPort is the default port of the server.
	DefaultPort = ":5000"

	// DefaultVerbose is whether verbose output should be enabled by default.
	DefaultVerbose = false
)

type context struct {
	adapter store.Adapter
	config  *jsonhttp.Config
}

type handle func(http.ResponseWriter, *http.Request, httprouter.Params, *context) (interface{}, error)

type handler struct {
	context *context
	handle  handle
}

func (h handler) serve(w http.ResponseWriter, r *http.Request, p httprouter.Params, _ *jsonhttp.Config) (interface{}, error) {
	return h.handle(w, r, p, h.context)
}

// Info is the info returned by the root route.
type Info struct {
	Adapter interface{} `json:"adapter"`
}

// New create an instance of a server.
func New(a store.Adapter, c *jsonhttp.Config) *jsonhttp.Server {
	s := jsonhttp.New(c)
	ctx := &context{a, c}

	s.Get("/", handler{ctx, root}.serve)
	s.Post("/segments", handler{ctx, saveSegment}.serve)
	s.Get("/segments/:linkHash", handler{ctx, getSegment}.serve)
	s.Delete("/segments/:linkHash", handler{ctx, deleteSegment}.serve)
	s.Get("/segments", handler{ctx, findSegments}.serve)
	s.Get("/maps", handler{ctx, getMapIDs}.serve)

	return s
}

func root(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c *context) (interface{}, error) {
	adapterInfo, err := c.adapter.GetInfo()
	if err != nil {
		return nil, err
	}

	return &Info{
		Adapter: adapterInfo,
	}, nil
}

func saveSegment(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c *context) (interface{}, error) {
	decoder := json.NewDecoder(r.Body)

	var s cs.Segment
	if err := decoder.Decode(&s); err != nil {
		return nil, jsonhttp.NewErrBadRequest("")
	}
	if err := s.Validate(); err != nil {
		return nil, jsonhttp.NewErrHTTP(err.Error(), http.StatusBadRequest)
	}
	if err := c.adapter.SaveSegment(&s); err != nil {
		return nil, err
	}

	return s, nil
}

func getSegment(w http.ResponseWriter, r *http.Request, p httprouter.Params, c *context) (interface{}, error) {
	linkHash, err := types.NewBytes32FromString(p.ByName("linkHash"))
	if err != nil {
		return nil, err
	}

	s, err := c.adapter.GetSegment(linkHash)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, jsonhttp.NewErrNotFound("")
	}

	return s, nil
}

func deleteSegment(w http.ResponseWriter, r *http.Request, p httprouter.Params, c *context) (interface{}, error) {
	linkHash, err := types.NewBytes32FromString(p.ByName("linkHash"))
	if err != nil {
		return nil, err
	}

	s, err := c.adapter.DeleteSegment(linkHash)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, jsonhttp.NewErrNotFound("")
	}

	return s, nil
}

func findSegments(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c *context) (interface{}, error) {
	filter, e := parseFilter(r)
	if e != nil {
		return nil, e
	}

	slice, err := c.adapter.FindSegments(filter)
	if err != nil {
		return nil, err
	}

	return slice, nil
}

func getMapIDs(w http.ResponseWriter, r *http.Request, _ httprouter.Params, c *context) (interface{}, error) {
	pagination, e := parsePagination(r)
	if e != nil {
		return nil, e
	}

	slice, err := c.adapter.GetMapIDs(pagination)
	if err != nil {
		return nil, err
	}

	return slice, nil
}

func parseFilter(r *http.Request) (*store.Filter, error) {
	pagination, err := parsePagination(r)
	if err != nil {
		return nil, err
	}

	var (
		mapID           = r.URL.Query().Get("mapId")
		prevLinkHashStr = r.URL.Query().Get("prevLinkHash")
		tagsStr         = r.URL.Query().Get("tags")
		prevLinkHash    *types.Bytes32
		tags            []string
	)

	if prevLinkHashStr != "" {
		prevLinkHash, err = types.NewBytes32FromString(prevLinkHashStr)
		if err != nil {
			return nil, newErrPrevLinkHash("")
		}
	}

	if tagsStr != "" {
		spaceTags := strings.Split(tagsStr, " ")
		for _, t := range spaceTags {
			tags = append(tags, strings.Split(t, "+")...)
		}
	}

	return &store.Filter{
		Pagination:   *pagination,
		MapID:        mapID,
		PrevLinkHash: prevLinkHash,
		Tags:         tags,
	}, nil
}

func parsePagination(r *http.Request) (*store.Pagination, error) {
	var err error

	offsetstr := r.URL.Query().Get("offset")
	offset := 0
	if offsetstr != "" {
		if offset, err = strconv.Atoi(offsetstr); err != nil || offset < 0 {
			return nil, newErrOffset("")
		}
	}

	limitstr := r.URL.Query().Get("limit")
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
