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
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
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
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return cs.SegmentSlice{}, nil }

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.NoError(t, err)
			assert.Len(t, validators, 0)
		})

		t.Run("No validators found for process", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
				if len(filter.Tags) > 1 {
					return cs.SegmentSlice{}, nil
				}
				link := cstesting.NewLinkBuilder().
					WithProcess(validation.GovernanceProcessName).
					WithTags(validation.ValidatorTag, "process").
					Build()
				return cs.SegmentSlice{link.Segmentify()}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.EqualError(t, err, validation.ErrValidatorNotFound.Error())
			assert.Nil(t, validators)
		})

		t.Run("Incomplete governance segment", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
				if len(filter.Tags) > 1 {
					link := cstesting.NewLinkBuilder().WithState(map[string]interface{}{
						"pki": "test",
					}).Build()
					return cs.SegmentSlice{link.Segmentify()}, nil
				}
				return cs.SegmentSlice{cstesting.NewLinkBuilder().WithTags("1", "2").Build().Segmentify()}, nil
			}

			s := validation.NewStore(a, &validation.Config{})
			validators, err := s.GetValidators(ctx)
			assert.EqualError(t, err, validation.ErrBadGovernanceSegment.Error())
			assert.Nil(t, validators)
		})

		t.Run("Bad governance segment format", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
				if len(filter.Tags) > 1 {
					link := cstesting.NewLinkBuilder().WithState(map[string]interface{}{
						"pki":   "test",
						"types": 1,
					}).Build()
					return cs.SegmentSlice{link.Segmentify()}, nil
				}
				return cs.SegmentSlice{cstesting.NewLinkBuilder().WithTags("1", "2").Build().Segmentify()}, nil
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

	t.Run("TestUpdateValidator", func(t *testing.T) {
		process := "auction"
		auctionPKI, _ := testutils.LoadPKI([]byte(testutils.ValidAuctionJSONPKIConfig))
		auctionTypes, _ := testutils.LoadTypes([]byte(testutils.ValidAuctionJSONTypesConfig))

		t.Run("Fail to fetch segments", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) { return nil, errors.New("error") }

			s := validation.NewStore(a, &validation.Config{})
			err := s.UpdateValidator(ctx, process, validation.RulesSchema{
				Types: auctionTypes,
				PKI:   *auctionPKI,
			})
			assert.EqualError(t, err, "Cannot retrieve governance segments: error")
		})

		t.Run("Creates new validator", func(t *testing.T) {
			a := dummystore.New(nil)

			s := validation.NewStore(a, &validation.Config{
				PluginsPath: pluginsPath,
			})
			err := s.UpdateValidator(ctx, process, validation.RulesSchema{
				Types: auctionTypes,
				PKI:   *auctionPKI,
			})

			validators, err := s.GetValidators(ctx)
			assert.NoError(t, err)
			require.Len(t, validators, 1)
			assert.Len(t, validators[0], 6)

			segments, err := a.FindSegments(ctx, &store.SegmentFilter{
				Pagination: store.Pagination{Limit: 1},
				Process:    validation.GovernanceProcessName,
				Tags:       []string{process, validation.ValidatorTag},
			})
			assert.NoError(t, err)
			require.Len(t, segments, 1)
		})

		t.Run("Fails to create new validator", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) {
				return cs.SegmentSlice{cstesting.RandomSegment()}, nil
			}

			a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) { return nil, errors.New("error") }

			s := validation.NewStore(a, &validation.Config{})
			err := s.UpdateValidator(ctx, process, validation.RulesSchema{
				Types: auctionTypes,
				PKI:   *auctionPKI,
			})
			assert.EqualError(t, err, "cannot create link for process governance auction: error")
		})

		t.Run("Updates an existing validator", func(t *testing.T) {
			a := dummystore.New(nil)
			s := validation.NewStore(a, &validation.Config{})

			// Insert an "auction" governance process in the store.
			populateStoreWithValidData(t, a)
			checkLastValidatorPriority(t, a, process, 1)

			updatedAuctionPKI, _ := testutils.LoadPKI([]byte(strings.Replace(testutils.ValidAuctionJSONPKIConfig, "alice", "jérome", -1)))

			err := s.UpdateValidator(ctx, process, validation.RulesSchema{
				Types: auctionTypes,
				PKI:   *updatedAuctionPKI,
			})
			require.NoError(t, err)

			// Make sure the priority has been increased.
			checkLastValidatorPriority(t, a, process, 2)
		})

		t.Run("Fails to update an existing validator", func(t *testing.T) {
			chatPKI := json.RawMessage(testutils.ValidChatJSONPKIConfig)

			a := new(storetesting.MockAdapter)
			a.MockFindSegments.Fn = func(*store.SegmentFilter) (cs.SegmentSlice, error) {
				return cs.SegmentSlice{
					cstesting.NewLinkBuilder().
						WithState(map[string]interface{}{"types": auctionTypes, "pki": chatPKI}).
						Build().
						Segmentify()}, nil
			}
			a.MockCreateLink.Fn = func(l *cs.Link) (*types.Bytes32, error) { return nil, errors.New("error") }

			s := validation.NewStore(a, &validation.Config{})
			err := s.UpdateValidator(ctx, process, validation.RulesSchema{
				Types: auctionTypes,
				PKI:   *auctionPKI,
			})
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
				link := cstesting.NewLinkBuilder().
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

func checkLastValidatorPriority(t *testing.T, a store.Adapter, process string, expected float64) {
	segs, err := a.FindSegments(context.Background(), &store.SegmentFilter{
		Pagination: store.Pagination{
			Offset: 0,
			Limit:  1,
		},
		Process: validation.GovernanceProcessName,
		Tags:    []string{process, validation.ValidatorTag},
	})
	assert.NoError(t, err, "FindSegment(governance) should sucess")
	require.Len(t, segs, 1, "The last validator config should be retrieved")
	assert.Equal(t, expected, segs[0].Link.Meta.Priority, "The last validator config should be retrieved")
}

func populateStoreWithValidData(t *testing.T, a store.LinkWriter) {
	auctionPKI, _ := testutils.LoadPKI([]byte(testutils.ValidAuctionJSONPKIConfig))
	auctionTypes, _ := testutils.LoadTypes([]byte(testutils.ValidAuctionJSONTypesConfig))
	link := createGovernanceLink("auction", auctionPKI, auctionTypes)
	hash, err := a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
	assert.NotNil(t, hash, "LinkHash should not be nil")

	auctionPKI, _ = testutils.LoadPKI([]byte(strings.Replace(testutils.ValidAuctionJSONPKIConfig, "alice", "charlie", -1)))
	link = createGovernanceLink("auction", auctionPKI, auctionTypes)
	link.Meta.PrevLinkHash = hash.String()
	link.Meta.Priority = 1.
	_, err = a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)

	chatPKI, _ := testutils.LoadPKI([]byte(testutils.ValidChatJSONPKIConfig))
	chatTypes, _ := testutils.LoadTypes([]byte(testutils.ValidChatJSONTypesConfig))
	link = createGovernanceLink("chat", chatPKI, chatTypes)
	_, err = a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
}

func createGovernanceLink(process string, pki *validators.PKI, types map[string]validation.TypeSchema) *cs.Link {
	state := make(map[string]interface{}, 0)

	state["pki"] = pki
	state["types"] = types

	link := cstesting.NewLinkBuilder().
		WithProcess(validation.GovernanceProcessName).
		WithTags(process, validation.ValidatorTag).
		WithState(state).
		Build()
	link.Meta.Priority = 0.
	return link
}
