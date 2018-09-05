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

package storetestcases

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

// CreateLinkFunc is a type for a function that creates a link for benchmarks.
type CreateLinkFunc func(b *testing.B, numLinks, i int) *chainscript.Link

// RandomLink creates a link with random data.
func RandomLink(b *testing.B, numLinks, i int) *chainscript.Link {
	l, err := chainscript.NewLinkBuilder("p", "m").
		WithData(chainscripttest.RandomString(24)).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	return l
}

// RandomLinkMapID is a CreateLinkFunc that creates a random link with map ID.
// The map ID will be one of ten possible values.
func RandomLinkMapID(b *testing.B, numLinks, i int) *chainscript.Link {
	l, err := chainscript.NewLinkBuilder("p", fmt.Sprintf("%d", i%10)).Build()
	if err != nil {
		b.Fatal(err)
	}

	return l
}

// RandomLinkPrevLinkHash is a CreateLinkFunc that creates a random link with
// previous link hash.
// The previous link hash will be one of ten possible values.
func RandomLinkPrevLinkHash(b *testing.B, numLinks, i int) *chainscript.Link {
	lh, _ := hex.DecodeString(fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10))
	l, err := chainscript.NewLinkBuilder("p", "m").
		WithParent(lh).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	return l
}

// RandomLinkTags is a CreateLinkFunc that creates a random link with tags.
// The tags will contain one of ten possible values.
func RandomLinkTags(b *testing.B, numLinks, i int) *chainscript.Link {
	l, err := chainscript.NewLinkBuilder("p", "m").
		WithTags(fmt.Sprintf("%d", i%10)).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	return l
}

// RandomLinkMapIDTags is a CreateLinkFunc that creates a random link with map
// ID and tags.
// The map ID will be one of ten possible values.
// The tags will contain one of ten possible values.
func RandomLinkMapIDTags(b *testing.B, numLinks, i int) *chainscript.Link {
	l, err := chainscript.NewLinkBuilder("p", fmt.Sprintf("%d", i%10)).
		WithTags(fmt.Sprintf("%d", i%10)).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	return l
}

// RandomLinkPrevLinkHashTags is a CreateLinkFunc that creates a random link
// with previous link hash and tags.
// The previous link hash will be one of ten possible values.
// The tags will contain one of ten possible values.
func RandomLinkPrevLinkHashTags(b *testing.B, numLinks, i int) *chainscript.Link {
	lh, _ := hex.DecodeString(fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10))
	l, err := chainscript.NewLinkBuilder("p", "m").
		WithParent(lh).
		WithTags(fmt.Sprintf("%d", i%10)).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	return l
}

// MapFilterFunc is a type for a function that creates a mapId filter for
// benchmarks.
type MapFilterFunc func(b *testing.B, numLinks, i int) *store.MapFilter

// RandomPaginationOffset is a a PaginationFunc that create a pagination with a random offset.
func RandomPaginationOffset(b *testing.B, numLinks, i int) *store.MapFilter {
	return &store.MapFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
	}
}

// FilterFunc is a type for a function that creates a filter for benchmarks.
type FilterFunc func(b *testing.B, numLinks, i int) *store.SegmentFilter

// RandomFilterOffset is a a FilterFunc that create a filter with a random
// offset.
func RandomFilterOffset(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
	}
}

// RandomFilterOffsetMapID is a a FilterFunc that create a filter with a random
// offset and map ID.
// The map ID will be one of ten possible values.
func RandomFilterOffsetMapID(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetMapIDs is a a FilterFunc that create a filter with a random
// offset and 2 map IDs.
// The map ID will be one of ten possible values.
func RandomFilterOffsetMapIDs(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10), fmt.Sprintf("%d", (i+1)%10)},
	}
}

// RandomFilterOffsetPrevLinkHash is a a FilterFunc that create a filter with a
// random offset and previous link hash.
// The previous link hash will be one of ten possible values.
func RandomFilterOffsetPrevLinkHash(b *testing.B, numLinks, i int) *store.SegmentFilter {
	prevLinkHash, _ := types.NewBytes32FromString(fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10))
	prevLinkHashStr := prevLinkHash.String()
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		PrevLinkHash: &prevLinkHashStr,
	}
}

// RandomFilterOffsetTags is a a FilterFunc that create a filter with a random
// offset and map ID.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetTags(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		Tags: []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetMapIDTags is a a FilterFunc that create a filter with a
// random offset and map ID and tags.
// The map ID will be one of ten possible values.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetMapIDTags(b *testing.B, numLinks, i int) *store.SegmentFilter {
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		MapIDs: []string{fmt.Sprintf("%d", i%10)},
		Tags:   []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}

// RandomFilterOffsetPrevLinkHashTags is a a FilterFunc that create a filter
// with a random offset and previous link hash and tags.
// The previous link hash will be one of ten possible values.
// The tags will be one of fifty possible combinations.
func RandomFilterOffsetPrevLinkHashTags(b *testing.B, numLinks, i int) *store.SegmentFilter {
	prevLinkHash, _ := types.NewBytes32FromString(fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i%10))
	prevLinkHashStr := prevLinkHash.String()
	return &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: rand.Int() % numLinks,
			Limit:  store.DefaultLimit,
		},
		PrevLinkHash: &prevLinkHashStr,
		Tags:         []string{fmt.Sprintf("%d", i%5), fmt.Sprintf("%d", i%10)},
	}
}
