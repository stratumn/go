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

// Package tmstore implements a store that saves all the segments in a
// tendermint app
package tmstore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/types"
	"github.com/stratumn/sdk/utils"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/events"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/sdk/jsonhttp"
)

const (
	// Name is the name set in the store's information.
	Name = "tm"

	// Description is the description set in the store's information.
	Description = "Stratumn TM Store"

	// DefaultEndpoint is the default Tendermint endpoint.
	DefaultEndpoint = "tcp://127.0.0.1:46657"

	// DefaultWsRetryInterval is the default interval between Tendermint Wbesocket connection tries.
	DefaultWsRetryInterval = 5 * time.Second
)

// TMStore is the type that implements github.com/stratumn/sdk/store.Adapter.
type TMStore struct {
	config        *Config
	didSaveChans  []chan *cs.Segment
	tmClient      client.Client
	stoppingWS    bool
	clientFactory func(endpoint string) client.Client
}

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// Endoint used to communicate with Tendermint core
	Endpoint string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	TMAppInfo   interface{} `json:"tmAppDescription"`
	Version     string      `json:"version"`
	Commit      string      `json:"commit"`
}

// New creates a new instance of a TMStore.
func New(config *Config) *TMStore {
	return NewFromClient(config, func(endpoint string) client.Client {
		return client.NewHTTP(endpoint, "/websocket")
	})
}

// NewFromClient creates a new instance of a TMStore with the given client.
func NewFromClient(config *Config, clientFactory func(endpoint string) client.Client) *TMStore {
	return &TMStore{config, nil, clientFactory(config.Endpoint), false, clientFactory}
}

// StartWebsocket starts the websocket client and wait for New Block events.
func (t *TMStore) StartWebsocket() error {
	t.stoppingWS = false
	if _, err := t.tmClient.Start(); err != nil {
		return err
	}
	eventType := tmtypes.EventStringNewBlock()
	t.tmClient.AddListenerForEvent("TMStore", eventType, func(msg events.EventData) {
		if err := t.notifyDidSaveChans(msg); err != nil {
			log.Error(err)
		}
	})
	log.Info("Connected to TMPoP")
	return nil
}

// RetryStartWebsocket starts the websocket client and wait for New Block events, it retries on errors.
func (t *TMStore) RetryStartWebsocket(interval time.Duration) error {
	return utils.Retry(func(attempt int) (retry bool, err error) {
		err = t.StartWebsocket()
		if err != nil {
			// the tendermint RPC HTTPClient does not handle well connection errors
			// we have to recreate the client in case of errors
			// (Check if it is still needed on Tendermint updates)
			t.tmClient = t.clientFactory(t.config.Endpoint)
			log.Infof("%v, retrying...", err)
			time.Sleep(interval)
		}
		return true, err
	}, 0)
}

// StopWebsocket stops the websocket client.
func (t *TMStore) StopWebsocket() {
	t.stoppingWS = true
	t.tmClient.Stop()
}

func (t *TMStore) notifyDidSaveChans(evt events.EventData) error {
	if evt == nil {
		log.Debug("Received empty websocket message")
		return nil
	}

	tmevt, ok := evt.(tmtypes.TMEventData)
	if !ok {
		log.Debug("Received wrong websocket message")
		return nil
	}

	newBlock, ok := tmevt.Unwrap().(tmtypes.EventDataNewBlock)

	for _, txBytes := range newBlock.Block.Data.Txs {
		tx := &tmpop.Tx{}

		if err := json.Unmarshal(txBytes, tx); err != nil {
			return err
		}

		for _, c := range t.didSaveChans {
			c <- tx.Segment
		}
	}

	return nil
}

// AddDidSaveChannel implements
// github.com/stratumn/sdk/fossilizer.Store.AddDidSaveChannel.
func (t *TMStore) AddDidSaveChannel(saveChan chan *cs.Segment) {
	t.didSaveChans = append(t.didSaveChans, saveChan)
}

// GetInfo implements github.com/stratumn/sdk/store.Adapter.GetInfo.
func (t *TMStore) GetInfo() (interface{}, error) {
	response, err := t.sendQuery(tmpop.GetInfo, nil)
	if err != nil {
		return nil, err
	}

	info := &tmpop.Info{}
	err = json.Unmarshal(response.Value, info)
	if err != nil {
		return nil, err
	}

	return &Info{
		Name:        Name,
		Description: Description,
		TMAppInfo:   info,
		Version:     t.config.Version,
		Commit:      t.config.Commit,
	}, nil
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (t *TMStore) SaveSegment(segment *cs.Segment) error {
	tx := &tmpop.Tx{
		TxType:  tmpop.SaveSegment,
		Segment: segment,
	}
	_, err := t.broadcastTx(tx)

	return err
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (t *TMStore) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	response, err := t.sendQuery(tmpop.GetSegment, linkHash)
	if err != nil {
		return
	}
	if response.Value == nil {
		return
	}

	segment = &cs.Segment{}
	err = json.Unmarshal(response.Value, segment)
	if err != nil {
		return
	}

	// Return nil when no segment has been found (and not an empty segment)
	if segment.IsEmpty() {
		segment = nil
	}
	return
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (t *TMStore) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	tx := &tmpop.Tx{
		TxType:   tmpop.DeleteSegment,
		LinkHash: linkHash,
	}
	val, err := t.broadcastTx(tx)
	if err != nil {
		return nil, err
	}

	if val != nil && val.DeliverTx.Data != nil {
		segment = &cs.Segment{}
		err = json.Unmarshal(val.DeliverTx.Data, segment)
		if err != nil {
			return nil, err
		}
	}

	return
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (t *TMStore) FindSegments(filter *store.SegmentFilter) (segmentSlice cs.SegmentSlice, err error) {
	response, err := t.sendQuery(tmpop.FindSegments, filter)
	if err != nil {
		return
	}

	err = json.Unmarshal(response.Value, &segmentSlice)
	if err != nil {
		return
	}

	var proofSlice [][]byte
	err = json.Unmarshal(response.Proof, &proofSlice)
	if err != nil {
		return
	}

	return
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (t *TMStore) GetMapIDs(filter *store.MapFilter) (ids []string, err error) {
	response, err := t.sendQuery(tmpop.GetMapIDs, filter)
	err = json.Unmarshal(response.Value, &ids)
	if err != nil {
		return
	}

	return
}

// NewBatch implements github.com/stratumn/sdk/store.Adapter.NewBatch.
func (t *TMStore) NewBatch() (store.Batch, error) {
	return NewBatch(t), nil
}

// SaveValue implements github.com/stratumn/sdk/store.Adapter.SaveValue.
func (t *TMStore) SaveValue(key, value []byte) error {
	tx := &tmpop.Tx{
		TxType: tmpop.SaveValue,
		Key:    key,
		Value:  value,
	}
	_, err := t.broadcastTx(tx)
	return err
}

// GetValue implements github.com/stratumn/sdk/store.Adapter.GetValue.
func (t *TMStore) GetValue(key []byte) (value []byte, err error) {
	response, err := t.sendQuery(tmpop.GetValue, key)
	if err != nil {
		return nil, err
	}
	return response.Value, nil
}

// DeleteValue implements github.com/stratumn/sdk/store.Adapter.DeleteValue.
func (t *TMStore) DeleteValue(key []byte) (value []byte, err error) {
	tx := &tmpop.Tx{
		TxType: tmpop.DeleteValue,
		Key:    key,
	}

	val, err := t.broadcastTx(tx)
	if err != nil {
		return
	}
	if val != nil && val.DeliverTx.Data != nil {
		value = val.DeliverTx.Data
	}

	return
}

func (t *TMStore) broadcastTx(tx *tmpop.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	result, err := t.tmClient.BroadcastTxCommit(txBytes)
	if err != nil {
		return nil, err
	}
	if result.CheckTx.IsErr() {
		if result.CheckTx.Code == tmpop.CodeTypeValidation {
			// TODO: this package should be HTTP unaware, so
			// we need a better way to pass error types.
			return nil, jsonhttp.NewErrBadRequest(result.CheckTx.Error())
		}
		return nil, fmt.Errorf(result.CheckTx.Error())
	}
	if result.DeliverTx.IsErr() {
		return nil, fmt.Errorf(result.DeliverTx.Error())
	}

	return result, nil
}

func (t *TMStore) sendQuery(name string, args interface{}) (res *abci.ResultQuery, err error) {
	query, err := tmpop.BuildQueryBinary(args)
	if err != nil {
		return
	}

	result, err := t.tmClient.ABCIQuery(name, query, true)
	if err != nil {
		return
	}
	if !result.ResultQuery.Code.IsOK() {
		return res, fmt.Errorf("NOK Response from TMPop: %v", result.ResultQuery)
	}

	return result.ResultQuery, nil
}
