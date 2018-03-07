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

package elasticsearchstore

import (
	"encoding/hex"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/bufferedbatch"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
)

const (
	// Name is the name set in the store's information.
	Name = "elasticsearch"

	// Description is the description set in the store's information.
	Description = "Indigo's ElasticSearch Store"

	// DefaultURL is the default URL of the database.
	DefaultURL = "http://elasticsearch:9200"
)

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// The URL of the ElasticSearch database.
	URL string

	// Use sniffing feature of ElasticSearch.
	Sniffing bool
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// ESStore is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type ESStore struct {
	config     *Config
	eventChans []chan *store.Event
	client     *elastic.Client
}

type errorLogger struct{}

func (l errorLogger) Printf(format string, vars ...interface{}) {
	log.Errorf(format, vars...)
}

type infoLogger struct{}

func (l infoLogger) Printf(format string, vars ...interface{}) {
	log.Infof(format, vars...)
}

type debugLogger struct{}

func (l debugLogger) Printf(format string, vars ...interface{}) {
	log.Debugf(format, vars...)
}

// New creates a new instance of an ElasticSearch store.
func New(config *Config) (*ESStore, error) {

	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(config.URL),
		elastic.SetSniff(config.Sniffing),
		elastic.SetErrorLog(errorLogger{}),
		elastic.SetInfoLog(debugLogger{}),
		elastic.SetTraceLog(debugLogger{}),
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	esStore := &ESStore{
		config: config,
		client: client,
	}

	if err := esStore.createIndex(linksIndex); err != nil {
		return nil, err
	}

	if err := esStore.createIndex(evidencesIndex); err != nil {
		return nil, err
	}

	if err := esStore.createIndex(valuesIndex); err != nil {
		return nil, err
	}

	return esStore, nil
}

/********** Store adapter implementation **********/

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (es *ESStore) GetInfo() (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     es.config.Version,
		Commit:      es.config.Commit,
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel.
func (es *ESStore) AddStoreEventChannel(eventChan chan *store.Event) {
	es.eventChans = append(es.eventChans, eventChan)
}

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (es *ESStore) NewBatch() (store.Batch, error) {
	return bufferedbatch.NewBatch(es), nil
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (es *ESStore) CreateLink(link *cs.Link) (*types.Bytes32, error) {
	linkHash, err := es.createLink(link)
	if err != nil {
		return nil, err
	}

	linkEvent := store.NewSavedLinks(link)

	es.notifyEvent(linkEvent)

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
func (es *ESStore) AddEvidence(linkHash *types.Bytes32, evidence *cs.Evidence) error {
	if err := es.addEvidence(linkHash.String(), evidence); err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	es.notifyEvent(evidenceEvent)

	return nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/go-indigocore/store.Adapter.GetSegment.
func (es *ESStore) GetSegment(linkHash *types.Bytes32) (*cs.Segment, error) {
	link, err := es.getLink(linkHash.String())
	if err != nil || link == nil {
		return nil, err
	}
	return es.segmentify(link), nil
}

// FindSegments implements github.com/stratumn/go-indigocore/store.Adapter.FindSegments.
func (es *ESStore) FindSegments(filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	return es.findSegments(filter)
}

// GetMapIDs implements github.com/stratumn/go-indigocore/store.Adapter.GetMapIDs.
func (es *ESStore) GetMapIDs(filter *store.MapFilter) ([]string, error) {
	return es.getMapIDs(filter)
}

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
func (es *ESStore) GetEvidences(linkHash *types.Bytes32) (*cs.Evidences, error) {
	return es.getEvidences(linkHash.String())
}

/********** github.com/stratumn/go-indigocore/store.KeyValueStore implementation **********/

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (es *ESStore) SetValue(key, value []byte) error {
	hexKey := hex.EncodeToString(key)
	return es.setValue(hexKey, value)
}

// GetValue implements github.com/stratumn/go-indigocore/store.Adapter.GetValue.
func (es *ESStore) GetValue(key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	return es.getValue(hexKey)
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.Adapter.DeleteValue.
func (es *ESStore) DeleteValue(key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	return es.deleteValue(hexKey)

}
