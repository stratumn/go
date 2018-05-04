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

package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetesting"
	"github.com/stratumn/go-indigocore/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGovernance(t *testing.T) {

	pluginsPath = "testdata"

	t.Run("New", func(t *testing.T) {
		t.Run("Governance without file", func(t *testing.T) {
			var v Validator
			a := new(storetesting.MockAdapter)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{})
			assert.NoError(t, err, "Gouvernance is initialized by store")
			assert.NotNil(t, gov, "Gouvernance is initialized by store")

			v = gov.Current()

			assert.Nil(t, v, "No validator loaded")
		})

		t.Run("Governance without file but store", func(t *testing.T) {
			var v Validator
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{})
			assert.NoError(t, err, "Gouvernance is initialized by store")
			assert.NotNil(t, gov, "Gouvernance is initialized by store")

			v = gov.Current()
			assert.NotNil(t, v, "Validator loaded from store")
		})

		t.Run("Governance with valid file", func(t *testing.T) {
			var v Validator
			a := new(storetesting.MockAdapter)
			testFile := utils.CreateTempFile(t, ValidJSONConfig)
			defer os.Remove(testFile)

			gov, err := NewLocalGovernor(context.Background(), a, &Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			assert.NoError(t, err, "Gouvernance is initialized by file and store")
			assert.NotNil(t, gov, "Gouvernance is initialized by file and store")

			v = gov.Current()

			assert.NotNil(t, v, "Validator loaded from file")
		})

		t.Run("Governance with invalid file", func(t *testing.T) {
			var v Validator
			a := new(storetesting.MockAdapter)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{
				RulesPath: "governance_test.go",
			})
			assert.NoError(t, err, "Gouvernance is initialized by store")
			assert.NotNil(t, gov, "Gouvernance is initialized by store")

			v = gov.Current()

			assert.Nil(t, v, "No validator loaded")
		})

		t.Run("Governance with unexisting file", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{
				RulesPath: "foo/bar",
			})
			assert.Error(t, err, "Cannot initialize gouvernance with bad file")
			assert.Nil(t, gov, "Cannot initialize gouvernance with bad file")
		})

		t.Run("New validator uploaded at startup", func(t *testing.T) {
			var v Validator
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			checkLastValidatorPriority(t, a, "auction", 1.)
			testFile := utils.CreateTempFile(t, ValidJSONConfig)
			defer os.Remove(testFile)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			require.NotNil(t, gov, "Gouvernance is initialized by file and store")
			assert.NoError(t, err, "Validator updated")

			v = gov.Current()
			assert.NotNil(t, v, "Validator loaded from file")
			checkLastValidatorPriority(t, a, "auction", 2.)
		})
	})

	t.Run("ListenAndUpdate", func(t *testing.T) {
		t.Run("New validation file read on modification", func(t *testing.T) {
			var v Validator
			ctx := context.Background()
			validJSON := fmt.Sprintf(`{%s}`, ValidChatJSONConfig)
			a := dummystore.New(nil)
			testFile := utils.CreateTempFile(t, validJSON)
			defer os.Remove(testFile)
			gov, err := NewLocalGovernor(ctx, a, &Config{
				RulesPath: testFile,
			})
			require.NotNil(t, gov, "Gouvernance is initialized by file and store")
			go gov.ListenAndUpdate(ctx)
			waitValidator := gov.AddListener()
			v = <-waitValidator
			assert.NotNil(t, v, "Validator loaded from file")

			checkLastValidatorPriority(t, a, "chat", 0.)

			chatJSON := createValidatorJSON("chat",
				strings.Replace(ValidChatJSONPKIConfig, "Bob", "Dave", -1),
				ValidChatJSONTypesConfig)
			validJSON = fmt.Sprintf(`{%s}`, chatJSON)
			f, err := os.OpenFile(testFile, os.O_WRONLY, 0)
			require.NoErrorf(t, err, "cannot modify file %s", &Config{
				RulesPath: testFile,
			})
			defer f.Close()
			_, err = f.WriteString(validJSON)
			require.NoError(t, err, "tmpfile.WriteString()")

			v = <-waitValidator
			assert.NotNil(t, v, "Validator reloaded from file")

			checkLastValidatorPriority(t, a, "chat", 1.)
		})

		t.Run("closes subscribing channels on context cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			testFile := utils.CreateTempFile(t, "")
			defer os.Remove(testFile)
			gov, err := NewLocalGovernor(ctx, dummystore.New(nil), &Config{
				RulesPath: testFile,
			})
			require.NoError(t, err)
			go func() {
				require.EqualError(t, gov.ListenAndUpdate(ctx), context.Canceled.Error())
			}()
		})

		t.Run("return an error when no file watcher is set", func(t *testing.T) {
			ctx := context.Background()
			testFile := utils.CreateTempFile(t, "")
			defer os.Remove(testFile)
			gov, err := NewLocalGovernor(ctx, dummystore.New(nil), &Config{})
			require.NoError(t, err)
			go func() {
				require.EqualError(t, gov.ListenAndUpdate(ctx), ErrNoFileWatcher.Error())
			}()
		})
	})

	t.Run("Current", func(t *testing.T) {
		t.Run("returns the current validator set", func(t *testing.T) {
			ctx := context.Background()
			testFile := utils.CreateTempFile(t, "")
			defer os.Remove(testFile)
			gov, err := NewLocalGovernor(ctx, dummystore.New(nil), &Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			require.NoError(t, err)

			go gov.ListenAndUpdate(ctx)
			assert.Nil(t, gov.Current())

			ioutil.WriteFile(testFile, []byte(ValidJSONConfig), os.ModeTemporary)
			newValidator := <-gov.AddListener()
			assert.Equal(t, newValidator, gov.Current())
		})
	})

	t.Run("AddRemoveListener", func(t *testing.T) {
		t.Run("Adds a listener provided with the current valitor set", func(t *testing.T) {
			ctx := context.Background()
			testFile := utils.CreateTempFile(t, ValidJSONConfig)
			defer os.Remove(testFile)
			gov, err := NewLocalGovernor(ctx, dummystore.New(nil), &Config{
				RulesPath:   testFile,
				PluginsPath: pluginsPath,
			})
			require.NoError(t, err)
			select {
			case <-gov.AddListener():
				break
			case <-time.After(10 * time.Millisecond):
				t.Error("No validator in the channel")
			}
		})

		t.Run("Removes an unknown channel", func(t *testing.T) {
			ctx := context.Background()
			gov, _ := NewLocalGovernor(ctx, dummystore.New(nil), &Config{})
			gov.RemoveListener(make(chan Validator))
			gov.AddListener()
			gov.RemoveListener(make(chan Validator))
		})

		t.Run("Removes closes the channel", func(t *testing.T) {
			ctx := context.Background()
			gov, _ := NewLocalGovernor(ctx, dummystore.New(nil), &Config{})
			listener := gov.AddListener()
			gov.RemoveListener(listener)

			_, ok := <-listener
			assert.False(t, ok, "<-listener")
		})
	})

	t.Run("TestGetAllProcesses", func(t *testing.T) {
		t.Run("No process", func(t *testing.T) {
			a := new(storetesting.MockAdapter)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{})
			require.NoError(t, err, "Gouvernance is initialized by store")

			processes := gov.(*LocalGovernor).getAllProcesses(context.Background())
			assert.Empty(t, processes)
		})

		t.Run("2 processes", func(t *testing.T) {
			a := dummystore.New(nil)
			populateStoreWithValidData(t, a)
			gov, err := NewLocalGovernor(context.Background(), a, &Config{})
			require.NoError(t, err, "Gouvernance is initialized by store")

			processes := gov.(*LocalGovernor).getAllProcesses(context.Background())
			assert.Len(t, processes, 2)
		})

		t.Run("Lot of processeses", func(t *testing.T) {
			a := dummystore.New(nil)
			for i := 0; i < store.MaxLimit+42; i++ {
				link := cstesting.NewLinkBuilder().
					WithProcess(governanceProcessName).
					WithTags(fmt.Sprintf("p%d", i), validatorTag).
					Build()
				_, err := a.CreateLink(context.Background(), link)
				assert.NoErrorf(t, err, "Cannot insert link %+v", link)
			}
			gov, err := NewLocalGovernor(context.Background(), a, &Config{})
			require.NoError(t, err, "Gouvernance is initialized by store")
			processes := gov.(*LocalGovernor).getAllProcesses(context.Background())
			assert.Len(t, processes, store.MaxLimit+42)
		})
	})
}

func checkLastValidatorPriority(t *testing.T, a store.Adapter, process string, expected float64) {
	segs, err := a.FindSegments(context.Background(), &store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    governanceProcessName,
		Tags:       []string{process, validatorTag},
	})
	assert.NoError(t, err, "FindSegment(governance) should sucess")
	require.Len(t, segs, 1, "The last validator config should be retrieved")
	assert.Equal(t, expected, segs[0].Link.Meta.Priority, "The last validator config should be retrieved")
}

func populateStoreWithValidData(t *testing.T, a store.LinkWriter) {
	auctionPKI := json.RawMessage(ValidAuctionJSONPKIConfig)
	auctionTypes := json.RawMessage(ValidAuctionJSONTypesConfig)
	link := createGovernanceLink("auction", auctionPKI, auctionTypes)
	hash, err := a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
	assert.NotNil(t, hash, "LinkHash should not be nil")

	auctionPKI, _ = json.Marshal(strings.Replace(ValidAuctionJSONPKIConfig, "alice", "charlie", -1))
	link = createGovernanceLink("auction", auctionPKI, auctionTypes)
	link.Meta.PrevLinkHash = hash.String()
	link.Meta.Priority = 1.
	_, err = a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)

	chatPKI := json.RawMessage(ValidChatJSONPKIConfig)
	chatTypes := json.RawMessage(ValidChatJSONTypesConfig)
	link = createGovernanceLink("chat", chatPKI, chatTypes)
	_, err = a.CreateLink(context.Background(), link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
}

func createGovernanceLink(process string, pki, types json.RawMessage) *cs.Link {
	state := make(map[string]interface{}, 0)

	var unmarshalledData interface{}
	json.Unmarshal(pki, &unmarshalledData)
	state["pki"] = unmarshalledData
	json.Unmarshal(types, &unmarshalledData)
	state["types"] = unmarshalledData

	link := cstesting.NewLinkBuilder().
		WithProcess(governanceProcessName).
		WithTags(process, validatorTag).
		WithState(state).
		Build()
	link.Meta.Priority = 0.
	return link
}
