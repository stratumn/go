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

package validation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetesting"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-indigocore/validation"
	"github.com/stratumn/go-indigocore/validation/testutils"
	"github.com/stratumn/go-indigocore/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	ctx := context.Background()

	t.Run("TestGetValidators", func(t *testing.T) {
		t.Run("No processes", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) { return &types.PaginatedSegments{}, nil }

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.NoError(t, err)
			assert.Len(t, validators, 0)
		})

		t.Run("No validators found for process", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
				if len(filter.Tags) > 1 {
					return &types.PaginatedSegments{}, nil
				}

				segment := chainscripttest.NewLinkBuilder(t).
					WithProcess(validation.GovernanceProcessName).
					WithTags(validation.ValidatorTag, "process").
					Segmentify(t)

				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{segment},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.EqualError(t, err, validation.ErrValidatorNotFound.Error())
			assert.Nil(t, validators)
		})

		t.Run("Incomplete governance segment", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
				if len(filter.Tags) > 1 {
					segment := chainscripttest.NewLinkBuilder(t).
						WithData(t, map[string]interface{}{
							"pki": "test",
						}).
						Segmentify(t)

					return &types.PaginatedSegments{
						Segments:   types.SegmentSlice{segment},
						TotalCount: 1,
					}, nil
				}

				segment := chainscripttest.NewLinkBuilder(t).WithTags("1", "2").Segmentify(t)
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{segment},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.EqualError(t, err, validation.ErrBadGovernanceSegment.Error())
			assert.Nil(t, validators)
		})

		t.Run("Bad governance segment format", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
				if len(filter.Tags) > 1 {
					segment := chainscripttest.NewLinkBuilder(t).
						WithData(t, map[string]interface{}{
							"pki":   "test",
							"types": 1,
						}).
						Segmentify(t)

					return &types.PaginatedSegments{
						Segments:   types.SegmentSlice{segment},
						TotalCount: 1,
					}, nil
				}

				segment := chainscripttest.NewLinkBuilder(t).WithTags("1", "2").Segmentify(t)
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{segment},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.EqualError(t, err, validation.ErrBadGovernanceSegment.Error())
			assert.Nil(t, validators)
		})

		t.Run("Multiple validators", func(t *testing.T) {
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)

			s := validation.NewStore(a, &validation.Config{
				PluginsPath: pluginsPath,
			})
			validators, err := s.GetValidators(ctx)
			assert.NoError(t, err)
			require.Len(t, validators, 2)
		})
	})

	t.Run("TestLinkFromSchema", func(t *testing.T) {
		process := "auction"
		auctionPKI, _ := testutils.LoadPKI([]byte(testutils.ValidAuctionJSONPKIConfig))
		auctionTypes, _ := testutils.LoadTypes([]byte(testutils.ValidAuctionJSONTypesConfig))

		t.Run("Fails to fetch segments", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return nil, errors.New("error")
			}

			s := validation.NewStore(a, &validation.Config{})
			link, err := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   auctionPKI,
			})
			assert.EqualError(t, err, "Cannot retrieve governance segments: error")
			assert.Nil(t, link)
		})

		t.Run("Creates a segment without parent", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			link, err := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   auctionPKI,
			})
			assert.NoError(t, err)
			assert.Nil(t, link.Meta.PrevLinkHash)
			assert.Equal(t, 0., link.Meta.Priority)
			assert.Equal(t, validation.GovernanceProcessName, link.Meta.Process.Name)
			assert.NotNil(t, uuid.FromStringOrNil(link.Meta.MapId))
			assert.Contains(t, link.Meta.Tags, process)
			assert.Contains(t, link.Meta.Tags, validation.ValidatorTag)

			var metadata map[string]string
			require.NoError(t, link.StructurizeMetadata(&metadata))
			assert.Equal(t, process, metadata[validation.ProcessMetaKey])
		})

		t.Run("Creates a segment from a parent", func(t *testing.T) {
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": auctionPKI}).
				Segmentify(t)

			parentHash := parent.LinkHash()
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			link, err := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   auctionPKI,
			})
			assert.NoError(t, err)
			assert.Equal(t, []byte(parentHash), link.Meta.PrevLinkHash)
			assert.Equal(t, parent.Link.Meta.Priority+1, link.Meta.Priority)
			assert.Equal(t, parent.Link.Meta.MapId, link.Meta.MapId)
			assert.Equal(t, validation.GovernanceProcessName, link.Meta.Process.Name)
			assert.Contains(t, link.Meta.Tags, process)
			assert.Contains(t, link.Meta.Tags, validation.ValidatorTag)

			var metadata map[string]string
			require.NoError(t, link.StructurizeMetadata(&metadata))
			assert.Equal(t, process, metadata[validation.ProcessMetaKey])
		})
	})

	t.Run("TestUpdateValidator", func(t *testing.T) {
		process := "auction"
		auctionPKI, _ := testutils.LoadPKI([]byte(testutils.ValidAuctionJSONPKIConfig))
		auctionTypes, _ := testutils.LoadTypes([]byte(testutils.ValidAuctionJSONTypesConfig))
		updatedAuctionPKI, _ := testutils.LoadPKI([]byte(strings.Replace(testutils.ValidAuctionJSONPKIConfig, "alice", "jérome", -1)))

		t.Run("Fails to fetch segments", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return nil, errors.New("error")
			}

			s := validation.NewStore(a, &validation.Config{})
			err := s.UpdateValidator(ctx, createGovernanceLink(t, process, 0., auctionPKI, auctionTypes))
			assert.EqualError(t, err, "Cannot retrieve governance segments: error")
		})

		t.Run("Fails when priority does not match", func(t *testing.T) {
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": auctionPKI}).
				WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
				Segmentify(t)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			err := s.UpdateValidator(ctx, createGovernanceLink(t, process, 0., updatedAuctionPKI, auctionTypes))
			assert.EqualError(t, err, validation.ErrBadPriority.Error())
		})

		t.Run("Fails when prevLinkHash does not match", func(t *testing.T) {
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": auctionPKI}).
				WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
				Segmentify(t)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				PKI:   updatedAuctionPKI,
				Types: auctionTypes,
			})
			link.Meta.PrevLinkHash = []byte{42}
			err := s.UpdateValidator(ctx, link)
			assert.EqualError(t, err, validation.ErrBadPrevLinkHash.Error())
		})

		t.Run("Fails when mapID does not match", func(t *testing.T) {
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": auctionPKI}).
				WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
				Segmentify(t)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				PKI:   updatedAuctionPKI,
				Types: auctionTypes,
			})
			link.Meta.MapId = "bad"
			err := s.UpdateValidator(ctx, link)
			assert.EqualError(t, err, validation.ErrBadMapID.Error())
		})

		t.Run("Fails when process does not match", func(t *testing.T) {
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": auctionPKI}).
				WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
				Segmentify(t)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				PKI:   updatedAuctionPKI,
				Types: auctionTypes,
			})
			err := link.SetMetadata(map[string]string{validation.ProcessMetaKey: "bad"})
			require.NoError(t, err)

			err = s.UpdateValidator(ctx, link)
			assert.EqualError(t, err, validation.ErrBadProcess.Error())
		})

		t.Run("Creates new validator", func(t *testing.T) {
			a := dummystore.New(nil)

			s := validation.NewStore(a, &validation.Config{
				PluginsPath: pluginsPath,
			})
			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   auctionPKI,
			})
			err := s.UpdateValidator(ctx, link)

			validators, err := s.GetValidators(ctx)
			assert.NoError(t, err)
			require.Len(t, validators, 1)
			assert.Len(t, validators["auction"], 6)

			segments, err := a.FindSegments(ctx, &store.SegmentFilter{
				Pagination: store.Pagination{Limit: 1},
				Process:    validation.GovernanceProcessName,
				Tags:       []string{process, validation.ValidatorTag},
			})
			assert.NoError(t, err)
			require.Len(t, segments.Segments, 1)
		})

		t.Run("Fails to create new validator", func(t *testing.T) {
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": auctionPKI}).
				WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
				Segmentify(t)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}

			a.MockCreateLink.Fn = func(l *chainscript.Link) (chainscript.LinkHash, error) { return nil, errors.New("error") }

			s := validation.NewStore(a, &validation.Config{})
			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   updatedAuctionPKI,
			})
			err := s.UpdateValidator(ctx, link)
			assert.EqualError(t, err, "cannot create link for process governance auction: error")
		})

		t.Run("Updates an existing validator", func(t *testing.T) {
			a := dummystore.New(nil)
			s := validation.NewStore(a, &validation.Config{})

			// Insert an "auction" governance process in the store.
			populateStoreWithValidData(t, a)
			l := getLastValidator(t, a, process)
			assert.Equal(t, 1., l.Meta.Priority)

			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   updatedAuctionPKI,
			})
			err := s.UpdateValidator(ctx, link)
			require.NoError(t, err)

			// Make sure the priority has been increased.
			l = getLastValidator(t, a, process)
			assert.Equal(t, 2., l.Meta.Priority)
		})

		t.Run("Fails to update an existing validator", func(t *testing.T) {
			chatPKI := json.RawMessage(testutils.ValidChatJSONPKIConfig)
			parent := chainscripttest.NewLinkBuilder(t).
				WithData(t, map[string]interface{}{"types": auctionTypes, "pki": chatPKI}).
				WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
				Segmentify(t)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (*types.PaginatedSegments, error) {
				return &types.PaginatedSegments{
					Segments:   types.SegmentSlice{parent},
					TotalCount: 1,
				}, nil
			}
			a.MockCreateLink.Fn = func(l *chainscript.Link) (chainscript.LinkHash, error) { return nil, errors.New("error") }

			s := validation.NewStore(a, &validation.Config{})
			link, _ := s.LinkFromSchema(ctx, process, &validation.RulesSchema{
				Types: auctionTypes,
				PKI:   auctionPKI,
			})
			err := s.UpdateValidator(ctx, link)
			assert.EqualError(t, err, "cannot create link for process governance auction: error")
		})
	})

	t.Run("TestGetAllProcesses", func(t *testing.T) {
		t.Run("No process", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			s := validation.NewStore(a, &validation.Config{})

			processes := s.GetAllProcesses(context.Background())
			assert.Empty(t, processes)
		})

		t.Run("2 processes", func(t *testing.T) {
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			s := validation.NewStore(a, &validation.Config{})

			processes := s.GetAllProcesses(context.Background())
			assert.Len(t, processes, 2)
		})

		t.Run("Lot of processeses", func(t *testing.T) {
			a := dummystore.New(nil)
			for i := 0; i < store.MaxLimit+42; i++ {
				link := chainscripttest.NewLinkBuilder(t).
					WithProcess(validation.GovernanceProcessName).
					WithTags(fmt.Sprintf("p%d", i), validation.ValidatorTag).
					Build()
				_, err := a.CreateLink(context.Background(), link)
				assert.NoErrorf(t, err, "Cannot insert link %+v", link)
			}
			s := validation.NewStore(a, &validation.Config{})

			processes := s.GetAllProcesses(context.Background())
			assert.Len(t, processes, store.MaxLimit+42)
		})
	})
}

func getLastValidator(t *testing.T, a store.Adapter, process string) *chainscript.Link {
	segs, err := a.FindSegments(context.Background(), &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: 0,
			Limit:  1,
		},
		Process: validation.GovernanceProcessName,
		Tags:    []string{process, validation.ValidatorTag},
	})
	assert.NoError(t, err, "FindSegment(governance) should sucess")
	require.NotNil(t, segs, "Validator configs should be retrieved")
	require.Len(t, segs.Segments, 1, "The last validator config should be retrieved")
	return segs.Segments[0].Link
}

func populateStoreWithValidData(t *testing.T, a store.LinkWriter) {
	auctionPKI, _ := testutils.LoadPKI([]byte(testutils.ValidAuctionJSONPKIConfig))
	auctionTypes, _ := testutils.LoadTypes([]byte(testutils.ValidAuctionJSONTypesConfig))
	link := createGovernanceLink(t, "auction", 0., auctionPKI, auctionTypes)
	hash, err := a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
	assert.NotNil(t, hash, "LinkHash should not be nil")

	auctionPKI, _ = testutils.LoadPKI([]byte(strings.Replace(testutils.ValidAuctionJSONPKIConfig, "alice", "charlie", -1)))
	link = createGovernanceLink(t, "auction", 0., auctionPKI, auctionTypes)
	link.Meta.PrevLinkHash = hash
	link.Meta.Priority = 1.
	_, err = a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)

	chatPKI, _ := testutils.LoadPKI([]byte(testutils.ValidChatJSONPKIConfig))
	chatTypes, _ := testutils.LoadTypes([]byte(testutils.ValidChatJSONTypesConfig))
	link = createGovernanceLink(t, "chat", 0., chatPKI, chatTypes)
	_, err = a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
}

func createGovernanceLink(t *testing.T, process string, priority float64, pki *validators.PKI, types map[string]validation.TypeSchema) *chainscript.Link {
	state := make(map[string]interface{}, 0)
	state["pki"] = pki
	state["types"] = types

	return chainscripttest.NewLinkBuilder(t).
		WithProcess(validation.GovernanceProcessName).
		WithTags(process, validation.ValidatorTag).
		WithData(t, state).
		WithPriority(priority).
		WithMetadata(t, map[string]string{validation.ProcessMetaKey: process}).
		Build()
}
