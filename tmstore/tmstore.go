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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stratumn/go-indigocore/bufferedbatch"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/tmpop"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-indigocore/utils"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmcommon "github.com/tendermint/tmlibs/common"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/jsonhttp"
)

const (
	// Name is the name set in the store's information.
	Name = "tm"

	// Description is the description set in the store's information.
	Description = "Indigo's Tendermint Store"

	// DefaultEndpoint is the default Tendermint endpoint.
	DefaultEndpoint = "tcp://127.0.0.1:46657"

	// DefaultWsRetryInterval is the default interval between Tendermint Websocket connection attempts.
	DefaultWsRetryInterval = 5 * time.Second
)

// TMStore is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type TMStore struct {
	config          *Config
	ctx             context.Context
	tmEventChan     chan interface{}
	storeEventChans []chan *store.Event
	tmClient        client.Client
}

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string
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
func New(config *Config, tmClient client.Client) *TMStore {
	return &TMStore{
		config:   config,
		ctx:      context.Background(),
		tmClient: tmClient,
	}
}

// StartWebsocket starts the websocket client and wait for New Block events.
func (t *TMStore) StartWebsocket() error {
	if !t.tmClient.IsRunning() {
		if err := t.tmClient.Start(); err != nil && err != tmcommon.ErrAlreadyStarted {
			return err
		}
	}

	// TMPoP notifies us of store events that we forward to clients
	t.tmEventChan = make(chan interface{}, 10)
	go func() {
		for {
			_, ok := <-t.tmEventChan
			if !ok {
				break
			}

			go t.notifyStoreChans()
		}
	}()

	newBlocksQuery := fmt.Sprintf("%s='%s'", tmtypes.EventTypeKey, tmtypes.EventNewBlock)
	if err := t.tmClient.Subscribe(t.ctx, newBlocksQuery, t.tmEventChan); err != nil {
		return err
	}

	log.Info("Connected to TMPoP")
	return nil
}

// RetryStartWebsocket starts the websocket client and retries on errors.
func (t *TMStore) RetryStartWebsocket(interval time.Duration) error {
	return utils.Retry(func(attempt int) (retry bool, err error) {
		err = t.StartWebsocket()
		if err != nil {
			log.Infof("%v, retrying...", err)
			time.Sleep(interval)
		}
		return true, err
	}, 0)
}

// StopWebsocket stops the websocket client.
func (t *TMStore) StopWebsocket() error {
	// Note: no need to close t.tmEventChan, unsubscribing handles it
	if err := t.tmClient.UnsubscribeAll(t.ctx); err != nil {
		log.Warnf("Error unsubscribing to Tendermint events: %s", err.Error())
		return err
	}

	if t.tmClient.IsRunning() {
		if err := t.tmClient.Stop(); err != nil && err != tmcommon.ErrAlreadyStopped {
			log.Warnf("Error stopping Tendermint client: %s", err.Error())
			return err
		}
	}

	return nil
}

func (t *TMStore) notifyStoreChans() {
	var pendingEvents []*store.Event
	response, err := t.sendQuery(tmpop.PendingEvents, nil)
	if err != nil || response.Value == nil {
		log.Warn("Could not get pending events from TMPoP.")
	}

	err = json.Unmarshal(response.Value, &pendingEvents)
	if err != nil {
		log.Warn("TMPoP pending events could not be unmarshalled.")
	}

	for _, event := range pendingEvents {
		for _, c := range t.storeEventChans {
			c <- event
		}
	}
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel.
func (t *TMStore) AddStoreEventChannel(storeChan chan *store.Event) {
	t.storeEventChans = append(t.storeEventChans, storeChan)
}

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
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

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (t *TMStore) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return linkHash, err
	}

	tx := &tmpop.Tx{
		TxType:   tmpop.CreateLink,
		Link:     link,
		LinkHash: linkHash,
	}
	_, err = t.broadcastTx(tx)

	return linkHash, err
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
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

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	for _, c := range t.storeEventChans {
		c <- evidenceEvent
	}

	return nil
}

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
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

// GetSegment implements github.com/stratumn/go-indigocore/store.SegmentReader.GetSegment.
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

// FindSegments implements github.com/stratumn/go-indigocore/store.SegmentReader.FindSegments.
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

// GetMapIDs implements github.com/stratumn/go-indigocore/store.SegmentReader.GetMapIDs.
func (t *TMStore) GetMapIDs(filter *store.MapFilter) (ids []string, err error) {
	response, err := t.sendQuery(tmpop.GetMapIDs, filter)
	err = json.Unmarshal(response.Value, &ids)
	if err != nil {
		return
	}

	return
}

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (t *TMStore) NewBatch() (store.Batch, error) {
	return bufferedbatch.NewBatch(t), nil
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

func (t *TMStore) sendQuery(name string, args interface{}) (res *abci.ResponseQuery, err error) {
	query, err := tmpop.BuildQueryBinary(args)
	if err != nil {
		return
	}

	response, err := t.tmClient.ABCIQuery(name, query)
	if err != nil {
		return
	}
	if !response.Response.IsOK() {
		return res, fmt.Errorf("NOK Response from TMPop: %v", response.Response)
	}

	return &response.Response, nil
}
