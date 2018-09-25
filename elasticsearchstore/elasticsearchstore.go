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
	"context"
	"encoding/hex"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/bufferedbatch"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/validation/validators"
)

const (
	// Name is the name set in the store's information.
	Name = "elasticsearch"

	// Description is the description set in the store's information.
	Description = "Stratumn's ElasticSearch Store"

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

	// Logrus log level.
	LogLevel string
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// ESStore is the type that implements github.com/stratumn/go-core/store.Adapter.
type ESStore struct {
	config     *Config
	eventChans []chan *store.Event
	client     *elastic.Client
}

type errorLogger struct{}

func (l errorLogger) Printf(format string, vars ...interface{}) {
	log.Errorf(format, vars...)
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

	if config.LogLevel != "" {
		lvl, err := log.ParseLevel(config.LogLevel)

		if err != nil {
			return nil, err
		}

		log.SetLevel(lvl)
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	esStore := &ESStore{
		config: config,
		client: client,
	}

	if err := esStore.createLinksIndex(); err != nil {
		return nil, err
	}

	if err := esStore.createEvidencesIndex(); err != nil {
		return nil, err
	}

	if err := esStore.createValuesIndex(); err != nil {
		return nil, err
	}

	return esStore, nil
}

/********** Store adapter implementation **********/

// GetInfo implements github.com/stratumn/go-core/store.Adapter.GetInfo.
func (es *ESStore) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     es.config.Version,
		Commit:      es.config.Commit,
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-core/store.Adapter.AddStoreEventChannel.
func (es *ESStore) AddStoreEventChannel(eventChan chan *store.Event) {
	es.eventChans = append(es.eventChans, eventChan)
}

// NewBatch implements github.com/stratumn/go-core/store.Adapter.NewBatch.
func (es *ESStore) NewBatch(ctx context.Context) (store.Batch, error) {
	return bufferedbatch.NewBatch(ctx, es), nil
}

/********** Store writer implementation **********/

// CreateLink implements github.com/stratumn/go-core/store.LinkWriter.CreateLink.
func (es *ESStore) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	if err := link.Validate(ctx); err != nil {
		return nil, err
	}

	if err := validators.NewRefsValidator().Validate(ctx, es, link); err != nil {
		return nil, err
	}

	if link.Meta.OutDegree >= 0 {
		return nil, store.ErrOutDegreeNotSupported
	}

	linkHash, err := es.createLink(ctx, link)
	if err != nil {
		return nil, err
	}

	linkEvent := store.NewSavedLinks(link)

	es.notifyEvent(linkEvent)

	return linkHash, nil
}

// AddEvidence implements github.com/stratumn/go-core/store.EvidenceWriter.AddEvidence.
func (es *ESStore) AddEvidence(ctx context.Context, linkHash chainscript.LinkHash, evidence *chainscript.Evidence) error {
	if err := es.addEvidence(ctx, linkHash.String(), evidence); err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	es.notifyEvent(evidenceEvent)

	return nil
}

/********** Store reader implementation **********/

// GetSegment implements github.com/stratumn/go-core/store.Adapter.GetSegment.
func (es *ESStore) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	link, err := es.getLink(ctx, linkHash.String())
	if err != nil || link == nil {
		return nil, err
	}
	return es.segmentify(ctx, link), nil
}

// FindSegments implements github.com/stratumn/go-core/store.Adapter.FindSegments.
func (es *ESStore) FindSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	return es.findSegments(ctx, filter)
}

// GetMapIDs implements github.com/stratumn/go-core/store.Adapter.GetMapIDs.
func (es *ESStore) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	return es.getMapIDs(ctx, filter)
}

// GetEvidences implements github.com/stratumn/go-core/store.EvidenceReader.GetEvidences.
func (es *ESStore) GetEvidences(ctx context.Context, linkHash chainscript.LinkHash) (types.EvidenceSlice, error) {
	return es.getEvidences(ctx, linkHash.String())
}

/********** github.com/stratumn/go-core/store.KeyValueStore implementation **********/

// SetValue implements github.com/stratumn/go-core/store.KeyValueStore.SetValue.
func (es *ESStore) SetValue(ctx context.Context, key, value []byte) error {
	hexKey := hex.EncodeToString(key)
	return es.setValue(ctx, hexKey, value)
}

// GetValue implements github.com/stratumn/go-core/store.Adapter.GetValue.
func (es *ESStore) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	return es.getValue(ctx, hexKey)
}

// DeleteValue implements github.com/stratumn/go-core/store.Adapter.DeleteValue.
func (es *ESStore) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	hexKey := hex.EncodeToString(key)
	return es.deleteValue(ctx, hexKey)

}

/********** Search feature **********/

// SimpleSearchQuery searches through the store for segments matching query criteria
// using ES simple query string feature
func (es *ESStore) SimpleSearchQuery(ctx context.Context, query *SearchQuery) (*types.PaginatedSegments, error) {
	return es.simpleSearchQuery(ctx, query)
}

// MultiMatchQuery searches through the store for segments matching query criteria
// using ES multi match query
func (es *ESStore) MultiMatchQuery(ctx context.Context, query *SearchQuery) (*types.PaginatedSegments, error) {
	return es.multiMatchQuery(ctx, query)
}
