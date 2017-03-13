// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package tmstore implements a store that saves all the segments in a
// tendermint app
package tmstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/tmpop"
	"github.com/stratumn/sdk/types"
	"github.com/stratumn/sdk/utils"

	wire "github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	log "github.com/Sirupsen/logrus"
)

const (
	// Name is the name set in the store's information.
	Name = "tm"

	// Description is the description set in the store's information.
	Description = "Stratumn TM Store"

	// DefaultEndpoint is the default Tendermint endpoint
	DefaultEndpoint = "tcp://127.0.0.1:46657"

	// DefaultWsRetryInterval is the default interval between Tendermint Wbesocket connection tries
	DefaultWsRetryInterval = 5 * time.Second
)

// TMStore is the type that implements github.com/stratumn/sdk/store.Adapter.
type TMStore struct {
	config       *Config
	didSaveChans []chan *cs.Segment
	tmClient     *TMClient
	stoppingWS   bool
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

// New creates a new instance of a TMStore
func New(config *Config) *TMStore {
	client := NewTMClient(config.Endpoint)

	return &TMStore{config, nil, client, false}
}

// StartWebsocket starts the websocket client and wait for New Block events
func (t *TMStore) StartWebsocket() error {
	t.stoppingWS = false
	if err := t.tmClient.StartWebsocket(); err != nil {
		return err
	}
	eventType := tmtypes.EventStringNewBlock()
	t.tmClient.Subscribe(eventType)

	r, e, q := t.tmClient.GetEventChannels()

	for {
		select {
		case msg := <-r:
			if err := t.notifyDidSaveChans(msg); err != nil {
				log.Error(err)
			}
		case err := <-e:
			log.Error(err)
		case <-q:
			if t.stoppingWS {
				return nil
			}
			log.Error("Unexpected quit signal... Retrying")
			t.tmClient.StopWebsocket()
			t.RetryStartWebsocket(DefaultWsRetryInterval)
		}
	}
}

// RetryStartWebsocket starts the websocket client and wait for New Block events, it retries on errors
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

// StopWebsocket stops the websocket client
func (t *TMStore) StopWebsocket() {
	t.stoppingWS = true
	t.tmClient.StopWebsocket()
}

func (t *TMStore) notifyDidSaveChans(msg json.RawMessage) error {
	if msg == nil {
		log.Debug("Received empty websocket message")
		return nil
	}

	result, err := new(ctypes.TMResult), new(error)
	wire.ReadJSONPtr(result, msg, err)
	if *err != nil {
		return *err
	}

	var event *ctypes.ResultEvent
	switch (*result).(type) {
	case *ctypes.ResultEvent:
		event = (*result).(*ctypes.ResultEvent)
	default:
		return nil
	}

	if event.Name != "NewBlock" {
		return fmt.Errorf("Unexpected event received: %v", *event)
	}
	newBlock, _ := (event.Data).(tmtypes.EventDataNewBlock)

	for _, tx := range newBlock.Block.Data.Txs {
		segment := &cs.Segment{}

		if err := json.Unmarshal(tx, segment); err != nil {
			return err
		}

		for _, c := range t.didSaveChans {
			c <- segment
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
	info := &tmpop.Info{}
	err := t.sendQuery(tmpop.GetInfo, nil, info)

	return &Info{
		Name:        Name,
		Description: Description,
		TMAppInfo:   info,
		Version:     t.config.Version,
		Commit:      t.config.Commit,
	}, err
}

// SaveSegment implements github.com/stratumn/sdk/store.Adapter.SaveSegment.
func (t *TMStore) SaveSegment(segment *cs.Segment) error {
	tx, err := json.Marshal(segment)
	if err != nil {
		return err
	}

	if _, err = t.tmClient.BroadcastTxCommit(tx); err != nil {
		return err
	}

	return nil
}

// GetSegment implements github.com/stratumn/sdk/store.Adapter.GetSegment.
func (t *TMStore) GetSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	segment = &cs.Segment{}
	err = t.sendQuery(tmpop.GetSegment, linkHash, segment)

	// Return nil when no segment has been found (and not an empty segment)
	if segment.IsEmpty() {
		segment = nil
	}
	return
}

// DeleteSegment implements github.com/stratumn/sdk/store.Adapter.DeleteSegment.
func (t *TMStore) DeleteSegment(linkHash *types.Bytes32) (segment *cs.Segment, err error) {
	segment = &cs.Segment{}
	err = t.sendQuery(tmpop.DeleteSegment, linkHash, segment)

	// Return nil when no segment has been deleted (and not an empty segment)
	if segment.IsEmpty() {
		segment = nil
	}
	return
}

// FindSegments implements github.com/stratumn/sdk/store.Adapter.FindSegments.
func (t *TMStore) FindSegments(filter *store.Filter) (segmentSlice cs.SegmentSlice, err error) {
	segmentSlice = make(cs.SegmentSlice, 0)
	err = t.sendQuery(tmpop.FindSegments, filter, &segmentSlice)
	return
}

// GetMapIDs implements github.com/stratumn/sdk/store.Adapter.GetMapIDs.
func (t *TMStore) GetMapIDs(pagination *store.Pagination) (ids []string, err error) {
	ids = make([]string, 0)
	err = t.sendQuery(tmpop.GetMapIDs, pagination, &ids)
	return
}

func (t *TMStore) sendQuery(name string, args interface{}, result interface{}) error {
	query, err := tmpop.BuildQueryBinary(name, args)
	if err != nil {
		return err
	}

	res, err := t.tmClient.ABCIQuery(query)
	if err != nil {
		return err
	}
	if res.Result.IsErr() {
		return errors.New(res.Result.Error())
	}

	err = json.Unmarshal(res.Result.Data, result)

	return err
}
