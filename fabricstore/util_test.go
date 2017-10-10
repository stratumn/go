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

package fabricstore

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/testutil"
	"github.com/stratumn/sdk/types"
)

// NewTestClient returns a unit test FabricStore
func NewTestClient() *FabricStore {
	config := Config{
		ChannelID:   "mychannel",
		ChaincodeID: "pop",
		Version:     "0.1.0",
		Commit:      "00000000000000000000000000000000",
	}
	s := FabricStore{
		fabricClient: &MockClient{},
		config:       &config,
	}
	return &s
}

// MockClient implements subset of fabric-sdk-go ChannelClient interface
type MockClient struct{}

// ExecuteTx execute transaction
func (m *MockClient) ExecuteTx(req apitxn.ExecuteTxRequest) (tID apitxn.TransactionID, err error) {
	switch req.Fcn {
	case "SaveSegment":
		err = nArgsError(1, req.Args)
		if err != nil {
			return
		}

		segment := &cs.Segment{}
		err = json.Unmarshal(req.Args[0], segment)
		if err != nil {
			return
		}
	case "DeleteSegment":
		err = nArgsError(1, req.Args)
		if err != nil {
			return
		}

		_, err = types.NewBytes32FromString(string(req.Args[0]))
		if err != nil {
			return
		}
	case "SaveValue":
		err = nArgsError(2, req.Args)
		if err != nil {
			return
		}
		if len(req.Args) != 2 {
			err = errors.New("Expected exactly 2 arguments")
		}
	case "DeleteValue":
		err = nArgsError(1, req.Args)
		if err != nil {
			return
		}
	default:
		err = errors.New("Unknown execute tx function")
		return
	}

	tID = newTransactionID()
	return
}

// Query chaincode
func (m *MockClient) Query(req apitxn.QueryRequest) (result []byte, err error) {
	err = nArgsError(1, req.Args)
	if err != nil {
		return
	}

	switch req.Fcn {
	case "GetSegment":
		_, err = types.NewBytes32FromString(string(req.Args[0]))
		if err != nil {
			return
		}

		segment := cstesting.RandomSegment()
		result, err = json.Marshal(segment)

		return
	case "FindSegments":
		segmentFilter := &store.SegmentFilter{}
		err = json.Unmarshal(req.Args[0], segmentFilter)
		if err != nil {
			return
		}

		segment := cstesting.RandomSegment()
		segments := cs.SegmentSlice{segment}
		result, err = json.Marshal(segments)

		return
	case "GetMapIDs":
		mapFilter := &store.MapFilter{}
		err = json.Unmarshal(req.Args[0], mapFilter)
		if err != nil {
			return
		}

		mapIDs := []string{
			testutil.RandomString(24),
			testutil.RandomString(24),
		}
		result, err = json.Marshal(mapIDs)

		return
	case "GetValue":
		result = []byte("value")
		return
	}

	err = errors.New("Unknown query function")
	return
}

// QueryWithOpts allows the user to provide options for query (sync vs async, etc.)
func (m *MockClient) QueryWithOpts(request apitxn.QueryRequest, opt apitxn.QueryOpts) ([]byte, error) {
	return m.Query(request)
}

// ExecuteTxWithOpts allows the user to provide options for transaction execution (sync vs async, etc.)
func (m *MockClient) ExecuteTxWithOpts(request apitxn.ExecuteTxRequest, opt apitxn.ExecuteTxOpts) (apitxn.TransactionID, error) {
	return m.ExecuteTx(request)
}

// RegisterChaincodeEvent registers chain code event
// @param {chan bool} channel which receives event details when the event is complete
// @returns {object}  object handle that should be used to unregister
func (m *MockClient) RegisterChaincodeEvent(notify chan<- *apitxn.CCEvent, chainCodeID string, eventID string) apitxn.Registration {
	return nil
}

// UnregisterChaincodeEvent unregisters chain code event
func (m *MockClient) UnregisterChaincodeEvent(registration apitxn.Registration) error {
	return nil
}

// Close releases channel client resources (disconnects event hub etc.)
func (m *MockClient) Close() error {
	return nil
}

func newTransactionID() apitxn.TransactionID {
	return apitxn.TransactionID{
		ID:    testutil.RandomString(24),
		Nonce: []byte(testutil.RandomString(24)),
	}
}

func nArgsError(expected int, received [][]byte) error {
	if len(received) != expected {
		return errors.New("Expected exactly " + strconv.Itoa(expected) + " argument(s)")
	}
	return nil
}
