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

// Package fabricstore implements a store that saves all the segments in a
// Hyperledger Fabric distributed ledger
package fabricstore

import (
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

const (
	// Name is the name set in the store's information.
	Name = "fabric"

	// Description is the description set in the store's information.
	Description = "Stratumn Fabric Store"
)

func main() {}

// Config contains configuration optios for the store
type Config struct {
	// ChannelID used to send transactions
	ChannelID string

	// ChaincodeID used for transactions
	ChaincodeID string

	// ConfigFile path to netork configuration file (yaml)
	ConfigFile string

	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string
}

// FabricStore is the type that implements github.com/stratumn/sdk/store.Adapter.
type FabricStore struct {
	fabricClient apitxn.ChannelClient
	didSaveChans []chan *cs.Segment
	config       *Config
}

// Info is the info returned by GetInfo.
type Info struct {
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	FabricAppInfo interface{} `json:"fabricAppDescription"`
	Version       string      `json:"version"`
	Commit        string      `json:"commit"`
}

// New creates a new instance of FabricStore
func New(config *Config) (*FabricStore, error) {
	sdkOptions := fabapi.Options{
		ConfigFile: config.ConfigFile,
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		return nil, err
	}

	chClient, err := sdk.NewChannelClient(config.ChannelID, "Admin")
	if err != nil {
		return nil, err
	}

	return &FabricStore{
		fabricClient: chClient,
		config:       config,
	}, nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/fossilizer.Store.AddDidSaveChannel.
func (f *FabricStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	f.didSaveChans = append(f.didSaveChans, saveChan)
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (f *FabricStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:          Name,
		Description:   Description,
		FabricAppInfo: nil,
		Version:       f.config.Version,
		Commit:        f.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (f *FabricStore) SaveSegment(segment *cs.Segment) error {
	segmentBytes, err := json.Marshal(segment)
	if err != nil {
		return err
	}

	_, err = f.fabricClient.ExecuteTx(apitxn.ExecuteTxRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "SaveSegment",
		Args:        [][]byte{segmentBytes},
	})

	return err
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (f *FabricStore) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	response, err := f.fabricClient.Query(apitxn.QueryRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "GetSegment",
		Args:        [][]byte{[]byte(linkHash.String())},
	})
	if err != nil {
		return
	}
	if response == nil {
		return
	}

	segment = &cs.Segment{}
	err = json.Unmarshal(response, segment)
	if err != nil {
		return
	}

	if segment.IsEmpty() {
		segment = nil
	}
	return
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (f *FabricStore) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	segment, err = f.GetSegment(linkHash)
	if err != nil {
		return
	}

	_, err = f.fabricClient.ExecuteTx(apitxn.ExecuteTxRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "DeleteSegment",
		Args:        [][]byte{[]byte(linkHash.String())},
	})
	if err != nil {
		return
	}

	return
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (f *FabricStore) FindSegments(filter *store.SegmentFilter) (segmentSlice cs.SegmentSlice, err error) {
	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return
	}

	response, err := f.fabricClient.Query(apitxn.QueryRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "FindSegments",
		Args:        [][]byte{filterBytes},
	})
	if err != nil {
		return
	}

	err = json.Unmarshal(response, &segmentSlice)
	if err != nil {
		return
	}

	// This should be removed once limit and skip are implemented in fabric/couchDB
	segmentSlice = filter.Pagination.PaginateSegments(segmentSlice)

	return
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (f *FabricStore) GetMapIDs(filter *store.MapFilter) (ids []string, err error) {
	filterBytes, err := json.Marshal(filter)
	if err != nil {
		return
	}

	response, err := f.fabricClient.Query(apitxn.QueryRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "GetMapIDs",
		Args:        [][]byte{filterBytes},
	})
	if err != nil {
		return
	}

	err = json.Unmarshal(response, &ids)
	if err != nil {
		return
	}

	// This should be removed once limit and skip are implemented in fabric/couchDB
	ids = filter.Pagination.PaginateStrings(ids)

	return
}

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (f *FabricStore) NewBatch() (store.Batch, error) {
	return NewBatch(f), nil
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (f *FabricStore) SaveValue(key, value []byte) error {
	_, err := f.fabricClient.ExecuteTx(apitxn.ExecuteTxRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "SaveValue",
		Args:        [][]byte{key, value},
	})

	return err
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (f *FabricStore) GetValue(key []byte) (value []byte, err error) {
	response, err := f.fabricClient.Query(apitxn.QueryRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "GetValue",
		Args:        [][]byte{key},
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (f *FabricStore) DeleteValue(key []byte) (value []byte, err error) {
	value, err = f.GetValue(key)
	if err != nil {
		return nil, err
	}

	_, err = f.fabricClient.ExecuteTx(apitxn.ExecuteTxRequest{
		ChaincodeID: f.config.ChaincodeID,
		Fcn:         "DeleteValue",
		Args:        [][]byte{key},
	})

	return
}
