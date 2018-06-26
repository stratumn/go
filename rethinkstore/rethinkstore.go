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

// Package rethinkstore implements a store that saves all the segments in a
// RethinkDB database.
package rethinkstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/bufferedbatch"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/go-indigocore/utils"

	rethink "gopkg.in/dancannon/gorethink.v4"
)

func init() {
	rethink.SetTags("json", "gorethink")
}

const (
	// Name is the name set in the store's information.
	Name = "rethink"

	// Description is the description set in the store's information.
	Description = "Indigo's RethinkDB Store"

	// DefaultURL is the default URL of the database.
	DefaultURL = "rethinkdb:28015"

	// DefaultDB is the default database.
	DefaultDB = "test"

	// DefaultHard is whether to use hard durability by default.
	DefaultHard = true

	connectAttempts = 12
	connectTimeout  = 2 * time.Second
)

// Config contains configuration options for the store.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// The URL of the PostgreSQL database, such as "localhost:28015" order
	// "localhost:28015,localhost:28016,localhost:28017".
	URL string

	// The database name
	DB string

	// Whether to use hard durability.
	Hard bool
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
}

// Store is the type that implements github.com/stratumn/go-indigocore/store.Adapter.
type Store struct {
	config     *Config
	eventChans []chan *store.Event
	session    *rethink.Session
	db         rethink.Term
	links      rethink.Term
	evidences  rethink.Term
	values     rethink.Term
}

type linkWrapper struct {
	ID           []byte    `json:"id"`
	Content      *cs.Link  `json:"content"`
	Priority     float64   `json:"priority"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MapID        string    `json:"mapId"`
	PrevLinkHash []byte    `json:"prevLinkHash"`
	Tags         []string  `json:"tags,omitempty"`
	Process      string    `json:"process"`
}

type evidencesWrapper struct {
	ID        []byte        `json:"id"`
	Content   *cs.Evidences `json:"content"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type valueWrapper struct {
	ID    []byte `json:"id"`
	Value []byte `json:"value"`
}

// New creates an instance of a Store.
func New(config *Config) (*Store, error) {
	opts := rethink.ConnectOpts{Addresses: strings.Split(config.URL, ",")}

	var session *rethink.Session
	err := utils.Retry(func(attempt int) (bool, error) {
		var err error
		session, err = rethink.Connect(opts)
		if err != nil {
			log.WithFields(log.Fields{
				"attempt": attempt,
				"max":     connectAttempts,
			}).Warn(fmt.Sprintf("Unable to connect to RethinkDB, retrying in %v", connectTimeout))
			time.Sleep(connectTimeout)
			return true, err
		}
		return false, err
	}, connectAttempts)

	if err != nil {
		return nil, err
	}

	db := rethink.DB(config.DB)
	_, err = db.Wait(rethink.WaitOpts{
		Timeout: connectTimeout,
	}).Run(session)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Store{
		config:    config,
		session:   session,
		db:        db,
		links:     db.Table("links"),
		evidences: db.Table("evidences"),
		values:    db.Table("values"),
	}, nil
}

// AddStoreEventChannel implements github.com/stratumn/go-indigocore/store.Adapter.AddStoreEventChannel.
func (a *Store) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

// GetInfo implements github.com/stratumn/go-indigocore/store.Adapter.GetInfo.
func (a *Store) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// Rethink cannot retrieve nil slices, so we force putting empty slices in Link
func formatLink(link *cs.Link) {
	if link.Meta.Tags == nil {
		link.Meta.Tags = []string{}
	}
	if link.Meta.Inputs == nil {
		link.Meta.Inputs = []interface{}{}
	}
	if link.Meta.Refs == nil {
		link.Meta.Refs = []cs.SegmentReference{}
	}
	if link.Signatures == nil {
		link.Signatures = []*cs.Signature{}
	}
}

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (a *Store) CreateLink(ctx context.Context, link *cs.Link) (*types.Bytes32, error) {
	prevLinkHash := link.Meta.GetPrevLinkHash()

	formatLink(link)

	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}

	w := linkWrapper{
		ID:        linkHash[:],
		Content:   link,
		Priority:  link.Meta.Priority,
		UpdatedAt: time.Now().UTC(),
		MapID:     link.Meta.MapID,
		Tags:      link.Meta.Tags,
		Process:   link.Meta.Process,
	}

	if prevLinkHash != nil {
		w.PrevLinkHash = prevLinkHash[:]
	}

	if err := a.links.Get(linkHash).Replace(&w).Exec(a.session); err != nil {
		return nil, err
	}

	linkEvent := store.NewSavedLinks(link)

	for _, c := range a.eventChans {
		c <- linkEvent
	}

	return linkHash, nil
}

// GetSegment implements github.com/stratumn/go-indigocore/store.SegmentReader.GetSegment.
func (a *Store) GetSegment(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
	cur, err := a.links.Get(linkHash[:]).Run(a.session)

	if err != nil {
		return nil, err
	}
	defer cur.Close()

	var w linkWrapper
	if err := cur.One(&w); err != nil {
		if err == rethink.ErrEmptyResult {
			return nil, nil
		}
		return nil, err
	}

	segment := w.Content.Segmentify()
	if evidences, err := a.GetEvidences(ctx, segment.Meta.GetLinkHash()); evidences != nil && err == nil {
		segment.Meta.Evidences = *evidences
	}

	return segment, nil
}

// FindSegments implements github.com/stratumn/go-indigocore/store.SegmentReader.FindSegments.
func (a *Store) FindSegments(ctx context.Context, filter *store.SegmentFilter) (cs.SegmentSlice, error) {
	var prevLinkHash []byte
	q := a.links

	if filter.PrevLinkHash != nil {
		if prevLinkHashBytes, err := types.NewBytes32FromString(*filter.PrevLinkHash); prevLinkHashBytes != nil && err == nil {
			prevLinkHash = prevLinkHashBytes[:]
		}
		q = q.Between([]interface{}{
			prevLinkHash,
			rethink.MinVal,
		}, []interface{}{
			prevLinkHash,
			rethink.MaxVal,
		}, rethink.BetweenOpts{
			Index:      "prevLinkHashOrder",
			LeftBound:  "closed",
			RightBound: "closed",
		})
	}

	if len(filter.LinkHashes) > 0 {

		linkHashes, err := cs.NewLinkHashesFromStrings(filter.LinkHashes)
		if err != nil {
			return nil, err
		}

		ids := make([]interface{}, len(linkHashes))
		for i, v := range linkHashes {
			ids[i] = v
		}
		q = q.GetAll(ids...)
	}

	if mapIDs := filter.MapIDs; len(mapIDs) > 0 {
		ids := make([]interface{}, len(mapIDs))
		for i, v := range mapIDs {
			ids[i] = v
		}
		q = q.Filter(func(row rethink.Term) interface{} {
			return rethink.Expr(ids).Contains(row.Field("mapId"))
		})
	} else if prevLinkHash := filter.PrevLinkHash; prevLinkHash != nil {
		q = q.OrderBy(rethink.OrderByOpts{Index: "prevLinkHashOrder"})
	} else if linkHashes := filter.LinkHashes; len(linkHashes) > 0 {
		q = q.OrderBy(rethink.Asc("id"))
	} else if mapIDs := filter.MapIDs; len(mapIDs) > 0 {
		q = q.OrderBy(rethink.OrderByOpts{Index: rethink.Desc("mapIdOrder")})
	} else {
		q = q.OrderBy(rethink.OrderByOpts{Index: rethink.Desc("order")})
	}

	if process := filter.Process; len(process) > 0 {
		q = q.Filter(rethink.Row.Field("process").Eq(process))
	}

	if tags := filter.Tags; len(tags) > 0 {
		t := make([]interface{}, len(tags))
		for i, v := range tags {
			t[i] = v
		}
		q = q.Filter(rethink.Row.Field("tags").Contains(t...))
	}

	q = q.OuterJoin(a.evidences, func(a, b rethink.Term) rethink.Term {
		return a.Field("id").Eq(b.Field("id"))
	}).Map(func(row rethink.Term) interface{} {
		return map[string]interface{}{
			"link": row.Field("left").Field("content"),
			"meta": map[string]interface{}{
				"evidences": rethink.Branch(row.HasFields("right"), row.Field("right").Field("content"), cs.Evidences{}),
			},
		}
	})

	cur, err := q.Skip(filter.Offset).Limit(filter.Limit).Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	segments := make(cs.SegmentSlice, 0, filter.Limit)
	if err := cur.All(&segments); err != nil {
		return nil, err
	}
	for _, s := range segments {
		err = s.SetLinkHash()
		if err != nil {
			return nil, err
		}
	}

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/go-indigocore/store.SegmentReader.GetMapIDs.
func (a *Store) GetMapIDs(ctx context.Context, filter *store.MapFilter) ([]string, error) {
	q := a.links
	if process := filter.Process; len(process) > 0 {

		q = q.Between([]interface{}{
			process,
			rethink.MinVal,
		}, []interface{}{
			process,
			rethink.MaxVal,
		}, rethink.BetweenOpts{
			Index:      "processOrder",
			LeftBound:  "closed",
			RightBound: "closed",
		})
		q = q.OrderBy(rethink.OrderByOpts{Index: "processOrder"}).
			Distinct(rethink.DistinctOpts{Index: "processOrder"}).
			Map(func(row rethink.Term) interface{} {
				return row.AtIndex(1)
			})
	} else {
		q = q.Between(rethink.MinVal, rethink.MaxVal, rethink.BetweenOpts{
			Index: "mapId",
		}).
			OrderBy(rethink.OrderByOpts{Index: "mapId"}).
			Distinct(rethink.DistinctOpts{Index: "mapId"})
	}

	cur, err := q.Skip(filter.Pagination.Offset).Limit(filter.Limit).Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	mapIDs := []string{}
	if err = cur.All(&mapIDs); err != nil {
		return nil, err
	}

	return mapIDs, nil
}

// AddEvidence implements github.com/stratumn/go-indigocore/store.EvidenceWriter.AddEvidence.
func (a *Store) AddEvidence(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
	cur, err := a.evidences.Get(linkHash).Run(a.session)
	if err != nil {
		return err
	}
	defer cur.Close()

	var ew evidencesWrapper
	if err := cur.One(&ew); err != nil {
		if err != rethink.ErrEmptyResult {
			return err
		}
	}

	currentEvidences := ew.Content
	if currentEvidences == nil {
		currentEvidences = &cs.Evidences{}
	}

	if err := currentEvidences.AddEvidence(*evidence); err != nil {
		return err
	}

	w := evidencesWrapper{
		ID:        linkHash[:],
		Content:   currentEvidences,
		UpdatedAt: time.Now(),
	}
	if err := a.evidences.Get(linkHash).Replace(&w).Exec(a.session); err != nil {
		return err
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	for _, c := range a.eventChans {
		c <- evidenceEvent
	}
	return nil
}

// GetEvidences implements github.com/stratumn/go-indigocore/store.EvidenceReader.GetEvidences.
func (a *Store) GetEvidences(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
	cur, err := a.evidences.Get(linkHash).Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	var ew evidencesWrapper
	if err := cur.One(&ew); err != nil {
		if err == rethink.ErrEmptyResult {
			return &cs.Evidences{}, nil
		}
		return nil, err
	}
	return ew.Content, nil
}

// GetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.GetValue.
func (a *Store) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	cur, err := a.values.Get(key).Run(a.session)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	var w valueWrapper
	if err := cur.One(&w); err != nil {
		if err == rethink.ErrEmptyResult {
			return nil, nil
		}
		return nil, err
	}

	return w.Value, nil
}

// SetValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.SetValue.
func (a *Store) SetValue(ctx context.Context, key, value []byte) error {
	v := &valueWrapper{
		ID:    key,
		Value: value,
	}

	return a.values.Get(key).Replace(&v).Exec(a.session)
}

// DeleteValue implements github.com/stratumn/go-indigocore/store.KeyValueStore.DeleteValue.
func (a *Store) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	res, err := a.values.
		Get(key).
		Delete(rethink.DeleteOpts{ReturnChanges: true}).
		RunWrite(a.session)
	if err != nil {
		return nil, err
	}
	if res.Deleted < 1 {
		return nil, nil
	}
	b, err := json.Marshal(res.Changes[0].OldValue)
	if err != nil {
		return nil, err
	}

	var w valueWrapper
	if err := json.Unmarshal(b, &w); err != nil {
		return nil, err
	}

	return w.Value, nil
}

type rethinkBufferedBatch struct {
	*bufferedbatch.Batch
}

// CreateLink implements github.com/stratumn/go-indigocore/store.LinkWriter.CreateLink.
func (b *rethinkBufferedBatch) CreateLink(ctx context.Context, link *cs.Link) (*types.Bytes32, error) {
	formatLink(link)
	return b.Batch.CreateLink(ctx, link)
}

// NewBatch implements github.com/stratumn/go-indigocore/store.Adapter.NewBatch.
func (a *Store) NewBatch(ctx context.Context) (store.Batch, error) {
	bbBatch := bufferedbatch.NewBatch(ctx, a)
	if bbBatch == nil {
		return nil, errors.New("cannot create underlying batch")
	}
	return &rethinkBufferedBatch{Batch: bbBatch}, nil
}

// Create creates the database tables and indexes.
func (a *Store) Create() (err error) {
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}

	exists, err := a.Exists()
	if err != nil {
		return err
	} else if exists {
		return nil
	}

	tblOpts := rethink.TableCreateOpts{}
	if !a.config.Hard {
		tblOpts.Durability = "soft"
	}

	exec(a.db.TableCreate("links", tblOpts))
	exec(a.links.Wait())
	exec(a.links.IndexCreate("mapId"))
	exec(a.links.IndexWait("mapId"))
	exec(a.links.IndexCreateFunc("order", []interface{}{
		rethink.Row.Field("priority"),
		rethink.Row.Field("updatedAt"),
	}))
	exec(a.links.IndexWait("order"))
	exec(a.links.IndexCreateFunc("mapIdOrder", []interface{}{
		rethink.Row.Field("mapId"),
		rethink.Row.Field("priority"),
		rethink.Row.Field("updatedAt"),
	}))
	exec(a.links.IndexWait("mapIdOrder"))
	exec(a.links.IndexCreateFunc("prevLinkHashOrder", []interface{}{
		rethink.Row.Field("prevLinkHash"),
		rethink.Row.Field("priority"),
		rethink.Row.Field("updatedAt"),
	}))
	exec(a.links.IndexWait("prevLinkHashOrder"))
	exec(a.links.IndexCreateFunc("processOrder", []interface{}{
		rethink.Row.Field("process"),
		rethink.Row.Field("mapId"),
	}))
	exec(a.links.IndexWait("processOrder"))

	exec(a.db.TableCreate("evidences", tblOpts))
	exec(a.evidences.Wait())

	exec(a.db.TableCreate("values", tblOpts))
	exec(a.values.Wait())

	return err
}

// Drop drops the database tables and indexes.
func (a *Store) Drop() (err error) {
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}
	exec(a.db.TableDrop("links"))
	exec(a.db.TableDrop("evidences"))
	exec(a.db.TableDrop("values"))

	return
}

// Clean removes all documents from the tables.
func (a *Store) Clean() (err error) {
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}
	exec(a.links.Delete())
	exec(a.evidences.Delete())
	exec(a.values.Delete())

	return
}

// Exists returns whether the database tables exists.
func (a *Store) Exists() (bool, error) {
	cur, err := a.db.TableList().Run(a.session)
	if err != nil {
		return false, err
	}
	defer cur.Close()

	var name string
	for cur.Next(&name) {
		if name == "links" || name == "evidences" || name == "values" {
			return true, nil
		}
	}
	return false, nil
}
