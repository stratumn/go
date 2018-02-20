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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetesting"
	"github.com/stratumn/go-indigocore/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForUpdate(gov *GovernanceManager, v *Validator) error {
	return utils.Retry(func(attempt int) (retry bool, err error) {
		if gov.UpdateValidators(v) {
			return false, nil
		}
		time.Sleep(25 * time.Millisecond)
		return true, errors.New("missing validator")
	}, 20)
}

func TestGovernanceCreation(t *testing.T) {
	t.Run("Governance without file", func(t *testing.T) {
		var v Validator
		a := new(storetesting.MockAdapter)
		gov, err := NewGovernanceManager(a, "")
		assert.NoError(t, err, "Gouvernance is initialized by store")
		assert.NotNil(t, gov, "Gouvernance is initialized by store")

		err = waitForUpdate(gov, &v)
		assert.Nil(t, v, "No validator loaded")
		assert.Error(t, err, "No validator loaded")
	})

	t.Run("Governance without file but store", func(t *testing.T) {
		var v Validator
		a := dummystore.New(nil)
		populateStoreWithValidData(t, a)
		gov, err := NewGovernanceManager(a, "")
		assert.NoError(t, err, "Gouvernance is initialized by store")
		assert.NotNil(t, gov, "Gouvernance is initialized by store")

		err = waitForUpdate(gov, &v)
		assert.NotNil(t, v, "Validator loaded from store")
		assert.NoError(t, err, "Validator updated")
	})

	t.Run("Governance with valid file", func(t *testing.T) {
		var v Validator
		a := new(storetesting.MockAdapter)
		testFile := createTempFile(t, validJSONConfig)
		defer os.Remove(testFile)
		gov, err := NewGovernanceManager(a, testFile)
		assert.NoError(t, err, "Gouvernance is initialized by file and store")
		assert.NotNil(t, gov, "Gouvernance is initialized by file and store")

		err = waitForUpdate(gov, &v)
		assert.NotNil(t, v, "Validator loaded from file")
		assert.NoError(t, err, "Validator updated")
	})

	t.Run("Governance with invalid file", func(t *testing.T) {
		var v Validator
		a := new(storetesting.MockAdapter)
		gov, err := NewGovernanceManager(a, "governance_test.go")
		assert.NoError(t, err, "Gouvernance is initialized by store")
		assert.NotNil(t, gov, "Gouvernance is initialized by store")

		err = waitForUpdate(gov, &v)
		assert.Nil(t, v, "No validator loaded")
		assert.Error(t, err, "No validator loaded")
	})

	t.Run("Governance with unexisting file", func(t *testing.T) {
		a := new(storetesting.MockAdapter)
		gov, err := NewGovernanceManager(a, "/foo/bar")
		assert.Error(t, err, "Cannot initialize gouvernance with bad file")
		assert.Nil(t, gov, "Cannot initialize gouvernance with bad file")
	})
}

func checkLastValidatorPriority(t *testing.T, a store.Adapter, process string, expected float64) {
	segs, err := a.FindSegments(&store.SegmentFilter{
		Pagination: defaultPagination,
		Process:    governanceProcessName,
		Tags:       []string{process, validatorTag},
	})
	assert.NoError(t, err, "FindSegment(governance) should sucess")
	require.Len(t, segs, 1, "The last validator config should be retrieved")
	assert.Equal(t, expected, segs[0].Link.GetPriority(), "The last validator config should be retrieved")
}

func TestGovernanceUpdate(t *testing.T) {
	t.Run("New validator uploaded at startup", func(t *testing.T) {
		var v Validator
		a := dummystore.New(nil)
		populateStoreWithValidData(t, a)
		checkLastValidatorPriority(t, a, "auction", 1.)
		testFile := createTempFile(t, validJSONConfig)
		defer os.Remove(testFile)
		gov, err := NewGovernanceManager(a, testFile)
		require.NotNil(t, gov, "Gouvernance is initialized by file and store")

		err = waitForUpdate(gov, &v)
		assert.NotNil(t, v, "Validator loaded from file")
		assert.NoError(t, err, "Validator updated")
		checkLastValidatorPriority(t, a, "auction", 2.)
	})

	t.Run("New validation file read on modification", func(t *testing.T) {
		var v Validator
		validJSON := fmt.Sprintf(`{%s}`, validChatJSONConfig)
		a := dummystore.New(nil)
		testFile := createTempFile(t, validJSON)
		defer os.Remove(testFile)
		gov, err := NewGovernanceManager(a, testFile)
		require.NotNil(t, gov, "Gouvernance is initialized by file and store")

		err = waitForUpdate(gov, &v)
		assert.NotNil(t, v, "Validator loaded from file")
		assert.NoError(t, err, "Validator updated")
		checkLastValidatorPriority(t, a, "chat", 0.)

		chatJSON := createValidatorJSON("chat",
			strings.Replace(validChatJSONPKIConfig, "Bob", "Dave", -1),
			validChatJSONTypesConfig)
		validJSON = fmt.Sprintf(`{%s}`, chatJSON)
		f, err := os.OpenFile(testFile, os.O_WRONLY, 0)
		require.NoErrorf(t, err, "cannot modify file %s", testFile)
		defer f.Close()
		_, err = f.WriteString(validJSON)
		require.NoError(t, err, "tmpfile.WriteString()")

		err = waitForUpdate(gov, &v)
		assert.NotNil(t, v, "Validator reloaded from file")
		assert.NoError(t, err, "Validator reloaded from file")
		checkLastValidatorPriority(t, a, "chat", 1.)
	})
}

func TestGetAllProcesses(t *testing.T) {
	t.Run("No process", func(t *testing.T) {
		a := new(storetesting.MockAdapter)
		gov, err := NewGovernanceManager(a, "")
		require.NoError(t, err, "Gouvernance is initialized by store")
		processes := gov.getAllProcesses()
		assert.Empty(t, processes)
	})

	t.Run("2 processes", func(t *testing.T) {
		a := dummystore.New(nil)
		populateStoreWithValidData(t, a)
		gov, err := NewGovernanceManager(a, "")
		require.NoError(t, err, "Gouvernance is initialized by store")
		processes := gov.getAllProcesses()
		assert.Len(t, processes, 2)
	})

	t.Run("Lot of processeses", func(t *testing.T) {
		a := dummystore.New(nil)
		for i := 0; i < store.MaxLimit+42; i++ {
			link := cstesting.RandomLink()
			link.Meta["process"] = governanceProcessName
			link.Meta["tags"] = []interface{}{fmt.Sprintf("p%d", i), validatorTag}
			_, err := a.CreateLink(link)
			assert.NoErrorf(t, err, "Cannot insert link %+v", link)
		}
		gov, err := NewGovernanceManager(a, "")
		require.NoError(t, err, "Gouvernance is initialized by store")
		processes := gov.getAllProcesses()
		assert.Len(t, processes, store.MaxLimit+42)
	})
}

func populateStoreWithValidData(t *testing.T, a store.LinkWriter) {
	auctionPKI := json.RawMessage(validAuctionJSONPKIConfig)
	auctionTypes := json.RawMessage(validAuctionJSONTypesConfig)
	link := createGovernanceLink("auction", auctionPKI, auctionTypes)
	hash, err := a.CreateLink(link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
	assert.NotNil(t, hash, "LinkHash should not be nil")

	auctionPKI, _ = json.Marshal(strings.Replace(validAuctionJSONPKIConfig, "alice", "charlie", -1))
	link = createGovernanceLink("auction", auctionPKI, auctionTypes)
	link.Meta["prevLinkHash"] = hash.String()
	link.Meta["priority"] = 1.
	_, err = a.CreateLink(link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)

	chatPKI := json.RawMessage(validChatJSONPKIConfig)
	chatTypes := json.RawMessage(validChatJSONTypesConfig)
	link = createGovernanceLink("chat", chatPKI, chatTypes)
	_, err = a.CreateLink(link)
	assert.NoErrorf(t, err, "Cannot insert link %+v", link)
}

func createGovernanceLink(process string, pki, types json.RawMessage) *cs.Link {
	link := cstesting.RandomLink()
	link.Meta["process"] = governanceProcessName
	link.Meta["priority"] = 0.
	link.Meta["tags"] = []interface{}{process, validatorTag}
	link.State["pki"] = pki
	link.State["types"] = types
	return link
}
