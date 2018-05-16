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
	"strings"
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/cs/cstesting"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetesting"
	"github.com/stratumn/go-indigocore/validation"
	"github.com/stratumn/go-indigocore/validation/testutils"
	"github.com/stratumn/go-indigocore/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkManager(t *testing.T) {

	auctionPKI, _ := testutils.LoadPKI([]byte(strings.Replace(testutils.ValidChatJSONPKIConfig, "Bob", "Dave", -1)))
	auctionTypes, _ := testutils.LoadTypes([]byte(testutils.ValidChatJSONTypesConfig))

	t.Run("New", func(t *testing.T) {
		linkChan := make(chan *cs.Link)
		t.Run("Manager without chan", func(t *testing.T) {
			var v validators.Validator
			a := new(storetesting.MockAdapter)
			gov, err := validation.NewNetworkManager(context.Background(), a, nil, &validation.Config{})
			assert.NoError(t, err)
			assert.NotNil(t, gov)

			v = gov.Current()
			assert.Nil(t, v, "No validator loaded")
		})

		t.Run("Manager loads rules from store", func(t *testing.T) {
			var v validators.Validator
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			gov, err := validation.NewNetworkManager(context.Background(), a, linkChan, &validation.Config{
				PluginsPath: pluginsPath,
			})
			assert.NoError(t, err, "Gouvernance is initialized by store")
			require.NotNil(t, gov, "Gouvernance is initialized by store")

			v = gov.Current()
			assert.NotNil(t, v, "Validator loaded from store")
		})

		t.Run("Manager fails to load rules from store", func(t *testing.T) {
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

			gov, err := validation.NewNetworkManager(context.Background(), a, linkChan, &validation.Config{})
			assert.EqualError(t, err, "could not initialize network governor: could not find governance segments")
			assert.Nil(t, gov)
		})

	})

	t.Run("ListenAndUpdate", func(t *testing.T) {
		linkChan := make(chan *cs.Link)

		t.Run("Update rules in store when receiving new ones", func(t *testing.T) {
			var v validators.Validator
			linkChan := make(chan *cs.Link)
			ctx := context.Background()
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)

			gov, err := validation.NewNetworkManager(ctx, a, linkChan, &validation.Config{
				PluginsPath: pluginsPath,
			})
			assert.NoError(t, err)
			require.NotNil(t, gov)

			waitValidator := gov.Subscribe()
			go func() {
				assert.NoError(t, gov.ListenAndUpdate(ctx))
			}()

			v = <-waitValidator
			assert.NotNil(t, v, "Validator loaded from store")

			l := getLastValidator(t, a, "chat")
			assert.Equal(t, 0., l.Meta.Priority)

			go func() {
				parent := getLastValidator(t, a, "chat")
				parentHash, _ := parent.HashString()
				newRules := cstesting.NewLinkBuilder().
					WithMapID(parent.Meta.MapID).
					WithPrevLinkHash(parentHash).
					WithProcess(validation.GovernanceProcessName).
					WithTags(validation.ValidatorTag, "chat").
					WithMetadata(validation.ProcessMetaKey, "chat").
					WithPriority(1.).
					WithState(map[string]interface{}{"pki": auctionPKI, "types": auctionTypes}).
					Build()
				linkChan <- newRules
			}()

			v = <-waitValidator
			assert.NotNil(t, v, "Validator reloaded from file")

			l = getLastValidator(t, a, "chat")
			assert.Equal(t, 1., l.Meta.Priority)
		})

		t.Run("does not update rules if governance process name is missing", func(t *testing.T) {
			ctx := context.Background()
			a := dummystore.New(nil)
			gov, err := validation.NewNetworkManager(ctx, a, linkChan, &validation.Config{})
			assert.NoError(t, err)

			waitValidator := gov.Subscribe()
			go func() {
				assert.NoError(t, gov.ListenAndUpdate(ctx))
			}()

			go func() {
				linkChan <- cstesting.NewLinkBuilder().
					WithTags("process", validation.ValidatorTag).
					WithState(map[string]interface{}{"pki": auctionPKI, "types": auctionTypes}).
					WithMetadata(validation.ProcessMetaKey, "process").
					Build()
			}()

			select {
			case <-waitValidator:
				assert.Fail(t, "should not update validation rules")
			case <-time.After(15 * time.Millisecond):
				break
			}
			assert.Nil(t, gov.Current(), "Validator not loaded from file")
		})

		t.Run("does not update rules if validator tag is missing", func(t *testing.T) {
			ctx := context.Background()
			a := dummystore.New(nil)
			gov, err := validation.NewNetworkManager(ctx, a, linkChan, &validation.Config{})
			assert.NoError(t, err)

			waitValidator := gov.Subscribe()
			go func() {
				assert.NoError(t, gov.ListenAndUpdate(ctx))
			}()

			go func() {
				linkChan <- cstesting.NewLinkBuilder().
					WithProcess(validation.GovernanceProcessName).
					WithTag("process").
					WithState(map[string]interface{}{"pki": auctionPKI, "types": auctionTypes}).
					WithMetadata(validation.ProcessMetaKey, "process").
					Build()
			}()

			select {
			case <-waitValidator:
				assert.Fail(t, "should not update validation rules")
			case <-time.After(15 * time.Millisecond):
				break
			}
			assert.Nil(t, gov.Current(), "Validator not loaded from file")
		})

		t.Run("does not update rules if process meta data is missing", func(t *testing.T) {
			ctx := context.Background()
			a := dummystore.New(nil)
			gov, err := validation.NewNetworkManager(ctx, a, linkChan, &validation.Config{})
			assert.NoError(t, err)

			waitValidator := gov.Subscribe()
			go func() {
				assert.NoError(t, gov.ListenAndUpdate(ctx))
			}()

			go func() {
				linkChan <- cstesting.NewLinkBuilder().
					WithProcess(validation.GovernanceProcessName).
					WithTags("process", validation.ValidatorTag).
					WithState(map[string]interface{}{"pki": auctionPKI, "types": auctionTypes}).
					Build()
			}()

			select {
			case <-waitValidator:
				assert.Fail(t, "should not update validation rules")
			case <-time.After(15 * time.Millisecond):
				break
			}
			assert.Nil(t, gov.Current(), "Validator not loaded from file")
		})

		t.Run("does not update rules if governance link is badly formatted", func(t *testing.T) {
			ctx := context.Background()
			a := dummystore.New(nil)
			gov, err := validation.NewNetworkManager(ctx, a, linkChan, &validation.Config{})
			assert.NoError(t, err)

			waitValidator := gov.Subscribe()
			go func() {
				assert.NoError(t, gov.ListenAndUpdate(ctx))
			}()

			go func() {
				// PKI is missing
				linkChan <- cstesting.NewLinkBuilder().
					WithProcess(validation.GovernanceProcessName).
					WithTags("process", validation.ValidatorTag).
					WithState(map[string]interface{}{"types": auctionTypes}).
					WithMetadata(validation.ProcessMetaKey, "process").
					Build()
			}()

			select {
			case <-waitValidator:
				assert.Fail(t, "should not update validation rules")
			case <-time.After(15 * time.Millisecond):
				break
			}
			assert.Nil(t, gov.Current(), "Validator not loaded from file")
		})

		t.Run("closes subscribing channels on context cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			gov, err := validation.NewNetworkManager(ctx, dummystore.New(nil), linkChan, &validation.Config{})
			require.NoError(t, err)

			done := make(chan struct{})
			go func() {
				require.EqualError(t, gov.ListenAndUpdate(ctx), context.Canceled.Error())
				done <- struct{}{}
			}()
			cancel()
			<-done
		})

		t.Run("return an error when no network channel is set", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			gov, err := validation.NewNetworkManager(ctx, dummystore.New(nil), nil, &validation.Config{})
			require.NoError(t, err)
			go func() {
				assert.EqualError(t, gov.ListenAndUpdate(ctx), validation.ErrNoNetworkListener.Error())
				cancel()
			}()
			<-ctx.Done()
		})
	})

	t.Run("Current", func(t *testing.T) {
		linkChan := make(chan *cs.Link)
		t.Run("returns the current validator set", func(t *testing.T) {
			ctx := context.Background()
			a := dummystore.New(nil)
			gov, err := validation.NewNetworkManager(ctx, a, linkChan, &validation.Config{
				PluginsPath: pluginsPath,
			})
			require.NoError(t, err)

			go gov.ListenAndUpdate(ctx)
			assert.Nil(t, gov.Current())

			newValidator := gov.Subscribe()
			go func() {
				newRules := cstesting.NewLinkBuilder().
					WithProcess(validation.GovernanceProcessName).
					WithTags(validation.ValidatorTag, "chat").
					WithPrevLinkHash("").
					WithMetadata(validation.ProcessMetaKey, "chat").
					WithPriority(0.).
					WithState(map[string]interface{}{"pki": auctionPKI, "types": auctionTypes}).
					Build()
				linkChan <- newRules
			}()
			v := <-newValidator
			assert.Equal(t, v, gov.Current())
		})
	})
}
