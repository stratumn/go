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

package store

import (
	"reflect"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/types"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/testutil"
)

const (
	sliceSize = 100
)

var (
	prevLinkHashTestingValue      string
	linkHashTestingValue          string
	badLinkHashTestingValue       string
	emptyPrevLinkHashTestingValue = ""

	segmentSlice cs.SegmentSlice
	stringSlice  []string
)

func init() {
	prevLinkHashTestingValue = testutil.RandomHash().String()
	badLinkHashTestingValue = testutil.RandomHash().String()

	segmentSlice = make(cs.SegmentSlice, sliceSize)
	stringSlice = make([]string, sliceSize)
	for i := 0; i < sliceSize; i++ {
		segmentSlice[i] = cstesting.RandomSegment()
		stringSlice[i] = testutil.RandomString(10)
	}
}

func defaultTestingSegment() *cs.Segment {
	link := &cs.Link{
		Meta: map[string]interface{}{
			"prevLinkHash": prevLinkHashTestingValue,
			"process":      "TheProcess",
			"mapId":        "TheMapId",
			"tags":         []interface{}{"Foo", "Bar"},
			"priority":     42,
		},
	}
	return link.Segmentify()
}

func emptyPrevLinkHashTestingSegment() *cs.Segment {
	seg := defaultTestingSegment()
	delete(seg.Link.Meta, "prevLinkHash")
	return seg
}

func TestSegmentFilter_Match(t *testing.T) {
	type fields struct {
		Pagination   Pagination
		MapIDs       []string
		Process      string
		PrevLinkHash *string
		LinkHashes   []*types.Bytes32
		Tags         []string
	}
	type args struct {
		segment *cs.Segment
	}
	linkHashesSegment := defaultTestingSegment()
	linkHashesSegmentHash := linkHashesSegment.GetLinkHash()
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Null segment",
			fields: fields{},
			args:   args{nil},
			want:   false,
		},
		{
			name:   "Empty filter",
			fields: fields{},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Good mapId",
			fields: fields{MapIDs: []string{"TheMapId"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad mapId",
			fields: fields{MapIDs: []string{"AMapId"}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Good several mapIds",
			fields: fields{MapIDs: []string{"TheMapId", "SecondMapId"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Good process",
			fields: fields{Process: "TheProcess"},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad process",
			fields: fields{Process: "AProcess"},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Empty prevLinkHash ko",
			fields: fields{PrevLinkHash: &emptyPrevLinkHashTestingValue},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Empty prevLinkHash ok",
			fields: fields{PrevLinkHash: &emptyPrevLinkHashTestingValue},
			args:   args{emptyPrevLinkHashTestingSegment()},
			want:   true,
		},
		{
			name:   "Good prevLinkHash",
			fields: fields{PrevLinkHash: &prevLinkHashTestingValue},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad prevLinkHash",
			fields: fields{PrevLinkHash: &badLinkHashTestingValue},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "LinkHashes ok",
			fields: fields{LinkHashes: []*types.Bytes32{testutil.RandomHash(), linkHashesSegmentHash}},
			args:   args{linkHashesSegment},
			want:   true,
		},
		{
			name:   "LinkHashes ko",
			fields: fields{LinkHashes: []*types.Bytes32{testutil.RandomHash()}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "One tag",
			fields: fields{Tags: []string{"Foo"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Two tags",
			fields: fields{Tags: []string{"Foo", "Bar"}},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Only one good tag",
			fields: fields{Tags: []string{"Foo", "Baz"}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
		{
			name:   "Bad tag",
			fields: fields{Tags: []string{"Hello"}},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := SegmentFilter{
				Pagination:   tt.fields.Pagination,
				MapIDs:       tt.fields.MapIDs,
				Process:      tt.fields.Process,
				LinkHashes:   tt.fields.LinkHashes,
				PrevLinkHash: tt.fields.PrevLinkHash,
				Tags:         tt.fields.Tags,
			}
			if got := filter.Match(tt.args.segment); got != tt.want {
				t.Errorf("SegmentFilter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapFilter_Match(t *testing.T) {
	type fields struct {
		Pagination Pagination
		Process    string
	}
	type args struct {
		segment *cs.Segment
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Null segment",
			fields: fields{},
			args:   args{nil},
			want:   false,
		},
		{
			name:   "Empty filter",
			fields: fields{},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Good process",
			fields: fields{Process: "TheProcess"},
			args:   args{defaultTestingSegment()},
			want:   true,
		},
		{
			name:   "Bad process",
			fields: fields{Process: "AProcess"},
			args:   args{defaultTestingSegment()},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := MapFilter{
				Pagination: tt.fields.Pagination,
				Process:    tt.fields.Process,
			}
			if got := filter.Match(tt.args.segment); got != tt.want {
				t.Errorf("MapFilter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func defaultTestingPagination() Pagination {
	return Pagination{
		Offset: 0,
		Limit:  10,
	}
}
func TestPagination_PaginateSegments(t *testing.T) {
	type fields struct {
		Offset int
		Limit  int
	}
	type args struct {
		a cs.SegmentSlice
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   cs.SegmentSlice
	}{
		{
			name: "Nothing to paginate",
			fields: fields{
				Offset: 0,
				Limit:  2 * sliceSize,
			},
			args: args{segmentSlice},
			want: segmentSlice,
		},
		{
			name: "Paginate from beginning",
			fields: fields{
				Offset: 0,
				Limit:  sliceSize / 2,
			},
			args: args{segmentSlice},
			want: segmentSlice[:sliceSize/2],
		},
		{
			name: "Paginate from offset",
			fields: fields{
				Offset: 5,
				Limit:  sliceSize / 2,
			},
			args: args{segmentSlice},
			want: segmentSlice[5 : 5+sliceSize/2],
		},
		{
			name: "Paginate zero limit",
			fields: fields{
				Offset: 0,
				Limit:  0,
			},
			args: args{segmentSlice},
			want: cs.SegmentSlice{},
		},
		{
			name: "Paginate outer offset",
			fields: fields{
				Offset: 2 * sliceSize,
				Limit:  sliceSize,
			},
			args: args{segmentSlice},
			want: cs.SegmentSlice{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{
				Offset: tt.fields.Offset,
				Limit:  tt.fields.Limit,
			}
			if got := p.PaginateSegments(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pagination.PaginateSegments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPagination_PaginateStrings(t *testing.T) {
	type fields struct {
		Offset int
		Limit  int
	}
	type args struct {
		a []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Nothing to paginate",
			fields: fields{
				Offset: 0,
				Limit:  2 * sliceSize,
			},
			args: args{stringSlice},
			want: stringSlice,
		},
		{
			name: "Paginate from beginning",
			fields: fields{
				Offset: 0,
				Limit:  sliceSize / 2,
			},
			args: args{stringSlice},
			want: stringSlice[:sliceSize/2],
		},
		{
			name: "Paginate from offset",
			fields: fields{
				Offset: 5,
				Limit:  sliceSize / 2,
			},
			args: args{stringSlice},
			want: stringSlice[5 : 5+sliceSize/2],
		},
		{
			name: "Paginate zero limit",
			fields: fields{
				Offset: 0,
				Limit:  0,
			},
			args: args{stringSlice},
			want: []string{},
		},
		{
			name: "Paginate outer offset",
			fields: fields{
				Offset: 2 * sliceSize,
				Limit:  sliceSize,
			},
			args: args{stringSlice},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{
				Offset: tt.fields.Offset,
				Limit:  tt.fields.Limit,
			}
			if got := p.PaginateStrings(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pagination.PaginateStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
