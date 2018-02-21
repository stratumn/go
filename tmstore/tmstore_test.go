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

package tmstore

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/jsonhttp"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/store/storetestcases"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	"golang.org/x/crypto/ed25519"
)

const itPrivKey = "3t39DaJp54JXnBuBR31K889hqAFNms3V5U5cWqaY5VmGbG8T5z0/AZRIRVlDXRE9pM/lKS5NHrSkn4GHCqKHjw=="

var (
	tmstore *TMStore
)

func newTestTMStore() (store.Adapter, error) {
	tmstore = NewTestClient()
	err := tmstore.RetryStartWebsocket(DefaultWsRetryInterval)
	if err != nil {
		return nil, err
	}

	return tmstore, nil
}

func resetTMPop(_ store.Adapter) {
	ResetNode()
}

func updateValidatorRulesFile(t *testing.T, in, out string) {
	rulesInputFile, err := os.OpenFile(in, os.O_RDONLY, 0666)
	defer rulesInputFile.Close()
	content, err := ioutil.ReadAll(rulesInputFile)
	require.NoErrorf(t, err, "Cannot read validator rule file %s", in)
	err = ioutil.WriteFile(out, content, 0666)
	require.NoErrorf(t, err, "Cannot write validator rule file %s", out)

	if testTmpop != nil {
		tmInfo := testTmpop.Info(abci.RequestInfo{})
		testTmpop.BeginBlock(abci.RequestBeginBlock{
			Hash: tmInfo.LastBlockAppHash,
			Header: abci.Header{
				AppHash: tmInfo.LastBlockAppHash,
				Height:  tmInfo.LastBlockHeight,
			},
		})
		time.Sleep(500 * time.Millisecond)
	}
}

func TestTMStore(t *testing.T) {
	rulesFilename := filepath.Join("testdata", "rules.test.json")
	updateValidatorRulesFile(t, "/dev/null", rulesFilename)
	testConfig := &tmpop.Config{ValidatorFilename: rulesFilename}
	node := StartNode(testConfig)
	defer node.Wait()
	defer node.Stop()
	defer os.Remove(rulesFilename)

	t.Run("Store test cases", func(t *testing.T) {
		storetestcases.Factory{
			New:  newTestTMStore,
			Free: resetTMPop,
		}.RunStoreTests(t)
	})

	t.Run("Validation", func(t *testing.T) {
		tmstore.StartWebsocket()
		updateValidatorRulesFile(t, filepath.Join("testdata", "rules.json"), rulesFilename)

		var err error
		t.Run("Validation succeeds", func(t *testing.T) {
			l := cstesting.RandomLink()
			l.Meta["process"] = "testProcess"
			l.Meta["type"] = "init"
			l.State["string"] = "test"

			privBytes, _ := base64.StdEncoding.DecodeString(itPrivKey)
			ITPrivateKey := ed25519.PrivateKey(privBytes)
			l = cstesting.SignLinkWithKey(l, ITPrivateKey)

			_, err = tmstore.CreateLink(l)
			assert.NoError(t, err, "CreateLink() failed")
		})

		t.Run("Schema validation failed", func(t *testing.T) {
			l := cstesting.RandomLink()
			l.Meta["process"] = "testProcess"
			l.Meta["type"] = "init"
			l.State["string"] = 42

			_, err = tmstore.CreateLink(l)
			assert.Error(t, err, "A validation error is expected")

			errHTTP, ok := err.(jsonhttp.ErrHTTP)
			assert.True(t, ok, "Invalid error received: want ErrHTTP")
			assert.Equal(t, http.StatusBadRequest, errHTTP.Status())
		})

		t.Run("Signature validation failed", func(t *testing.T) {
			l := cstesting.RandomLink()
			l = cstesting.SignLink(l)
			l.Meta["process"] = "testProcess"
			l.Meta["type"] = "init"
			l.State["string"] = "test"

			_, err = tmstore.CreateLink(l)
			assert.Error(t, err, "A validation error is expected")

			errHTTP, ok := err.(jsonhttp.ErrHTTP)
			assert.True(t, ok, "Invalid error received: want ErrHTTP")
			assert.Equal(t, http.StatusBadRequest, errHTTP.Status())
		})

		t.Run("Validation rules update succeeds", func(t *testing.T) {
			l := cstesting.RandomLink()
			l.Meta["process"] = "testProcess"
			l.Meta["type"] = "action"
			l.State["string"] = "test"

			privBytes, _ := base64.StdEncoding.DecodeString(itPrivKey)
			ITPrivateKey := ed25519.PrivateKey(privBytes)
			l = cstesting.SignLinkWithKey(l, ITPrivateKey)

			_, err = tmstore.CreateLink(l)
			assert.NoError(t, err, "CreateLink() failed")

			updateValidatorRulesFile(t, filepath.Join("testdata", "rules.new.json"), rulesFilename)

			l = cstesting.RandomLink()
			l.Meta["process"] = "testProcess"
			l.Meta["type"] = "action"
			l.State["string"] = "test"

			l = cstesting.SignLinkWithKey(l, ITPrivateKey)

			_, err = tmstore.CreateLink(l)
			assert.Error(t, err, "CreateLink() should failed because signature is missing")
		})
	})

	// TestWebSocket tests how the web socket with Tendermint behaves
	t.Run("Websocket", func(t *testing.T) {
		t.Run("Start and stop websocket", func(t *testing.T) {
			err := tmstore.StartWebsocket()
			assert.NoError(t, err)

			err = tmstore.StopWebsocket()
			assert.NoError(t, err)
		})

		t.Run("Start websocket multiple times", func(t *testing.T) {
			err := tmstore.StartWebsocket()
			assert.NoError(t, err)

			err = tmstore.StartWebsocket()
			assert.NoError(t, err)

			err = tmstore.StopWebsocket()
			assert.NoError(t, err)
		})

		t.Run("Stop already stopped websocket", func(t *testing.T) {
			err := tmstore.StopWebsocket()
			assert.EqualError(t, err, "subscription not found")
		})
	})
}
