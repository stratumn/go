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
	"github.com/tendermint/tmlibs/events"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/sdk/jsonhttp"
)

const (
	// Name is the name set in the store's information.
	Name = "tm"

	// Description is the description set in the store's information.
	Description = "Stratumn TM Store"

	// DefaultEndpoint is the default Tendermint endpoint.
	DefaultEndpoint = "tcp://127.0.0.1:46657"

	// DefaultWsRetryInterval is the default interval between Tendermint Websocket connection attempts.
	DefaultWsRetryInterval = 5 * time.Second
)

// TMStore is the type that implements github.com/stratumn/sdk/store.AdapterV2.
type TMStore struct {
	config          *Config
	storeEventChans []chan *store.Event
	tmClient        client.Client
	stoppingWS      bool
	clientFactory   func(endpoint string) client.Client
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

	// TMPoP notifies us of store events that we forward to clients
	t.tmClient.AddListenerForEvent("TMStore", tmpop.StoreEvents, func(msg events.EventData) {
		go t.notifyStoreChans(msg)
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

func (t *TMStore) notifyStoreChans(msg events.EventData) {
	storeEvents, ok := msg.(tmpop.StoreEventsData)
	if !ok {
		log.Debug("Event could not be read as a list of store events")
	}

	for _, event := range storeEvents.StoreEvents {
		for _, c := range t.storeEventChans {
			c <- event
		}
	}
}

// AddStoreEventChannel implements github.com/stratumn/sdk/store.AdapterV2.AddStoreEventChannel.
func (t *TMStore) AddStoreEventChannel(storeChan chan *store.Event) {
	t.storeEventChans = append(t.storeEventChans, storeChan)
}

// GetInfo implements github.com/stratumn/sdk/store.AdapterV2.GetInfo.
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

// CreateLink implements github.com/stratumn/sdk/store.LinkWriter.CreateLink.
func (t *TMStore) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, err
	}

	tx := &tmpop.Tx{
		TxType: tmpop.CreateLink,
		Link:   link,
	}
	_, err = t.broadcastTx(tx)

	return linkHash, err
}

// AddEvidence implements github.com/stratumn/sdk/store.EvidenceWriter.AddEvidence.
func (t *TMStore) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
	// Adding an external evidence does not require consensus
	// So it will not go through a blockchain transaction, but will rather
	// be stored in TMPoP's store directly
	_, err := t.sendQuery(
		tmpop.AddEvidence,
		struct {
			LinkHash *types.Bytes32
			Evidence *cs.Evidence
		}{
			linkHash,
			evidence,
		})

	if err != nil {
		return err
	}

	for _, c := range t.storeEventChans {
		c <- &store.Event{
			EventType: store.SavedEvidence,
			Details:   evidence,
		}
	}

	return nil
}

// GetEvidences implements github.com/stratumn/sdk/store.EvidenceReader.GetEvidences.
func (t *TMStore) GetEvidences(linkHash *types.Bytes32) (evidences *cs.Evidences, err error) {
	evidences = &cs.Evidences{}
	response, err := t.sendQuery(tmpop.GetEvidences, linkHash)
	if err != nil {
		return
	}
	if response.Value == nil {
		return
	}

	err = json.Unmarshal(response.Value, evidences)
	if err != nil {
		return
	}

	return
}

// GetSegment implements github.com/stratumn/sdk/store.SegmentReader.GetSegment.
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

// FindSegments implements github.com/stratumn/sdk/store.SegmentReader.FindSegments.
func (t *TMStore) FindSegments(filter *store.SegmentFilter) (segmentSlice cs.SegmentSlice, err error) {
	response, err := t.sendQuery(tmpop.FindSegments, filter)
	if err != nil {
		return
	}

	err = json.Unmarshal(response.Value, &segmentSlice)
	if err != nil {
		return
	}

	return
}

// GetMapIDs implements github.com/stratumn/sdk/store.SegmentReader.GetMapIDs.
func (t *TMStore) GetMapIDs(filter *store.MapFilter) (ids []string, err error) {
	response, err := t.sendQuery(tmpop.GetMapIDs, filter)
	err = json.Unmarshal(response.Value, &ids)
	if err != nil {
		return
	}

	return
}

// NewBatchV2 implements github.com/stratumn/sdk/store.AdapterV2.NewBatchV2.
func (t *TMStore) NewBatchV2() (store.BatchV2, error) {
	return nil, nil
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

	result, err := t.tmClient.ABCIQuery(name, query)
	if err != nil {
		return
	}
	if !result.ResultQuery.Code.IsOK() {
		return res, fmt.Errorf("NOK Response from TMPop: %v", result.ResultQuery)
	}

	return result.ResultQuery, nil
}
