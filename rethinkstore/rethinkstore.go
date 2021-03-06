// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/bufferedbatch"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/util"

	rethink "gopkg.in/dancannon/gorethink.v4"
)

func init() {
	rethink.SetTags("json", "gorethink")
}

const (
	// Name is the name set in the store's information.
	Name = "rethink"

	// Description is the description set in the store's information.
	Description = "Stratumn's RethinkDB Store"

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

// Store is the type that implements github.com/stratumn/go-core/store.Adapter.
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
	ID           []byte            `json:"id"`
	Content      *chainscript.Link `json:"content"`
	Priority     float64           `json:"priority"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	MapID        string            `json:"mapId"`
	PrevLinkHash []byte            `json:"prevLinkHash"`
	Tags         []string          `json:"tags,omitempty"`
	Process      string            `json:"process"`
	Step         string            `json:"step"`
}

type evidencesWrapper struct {
	ID        []byte              `json:"id"`
	Content   types.EvidenceSlice `json:"content"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

type valueWrapper struct {
	ID    []byte `json:"id"`
	Value []byte `json:"value"`
}

// New creates an instance of a Store.
func New(config *Config) (*Store, error) {
	opts := rethink.ConnectOpts{Addresses: strings.Split(config.URL, ",")}

	var session *rethink.Session
	err := util.Retry(func(attempt int) (bool, error) {
		var err error
		session, err = rethink.Connect(opts)
		if err != nil {
			monitoring.LogEntry().WithFields(log.Fields{
				"attempt": attempt,
				"max":     connectAttempts,
			}).Warn(fmt.Sprintf("Unable to connect to RethinkDB, retrying in %v", connectTimeout))
			time.Sleep(connectTimeout)
			return true, err
		}
		return false, err
	}, connectAttempts)

	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not create rethinkstore")
	}

	db := rethink.DB(config.DB)
	_, err = db.Wait(rethink.WaitOpts{
		Timeout: connectTimeout,
	}).Run(session)

	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not create rethinkstore")
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

// AddStoreEventChannel implements github.com/stratumn/go-core/store.Adapter.AddStoreEventChannel.
func (a *Store) AddStoreEventChannel(eventChan chan *store.Event) {
	a.eventChans = append(a.eventChans, eventChan)
}

// GetInfo implements github.com/stratumn/go-core/store.Adapter.GetInfo.
func (a *Store) GetInfo(ctx context.Context) (interface{}, error) {
	return &Info{
		Name:        Name,
		Description: Description,
		Version:     a.config.Version,
		Commit:      a.config.Commit,
	}, nil
}

// Rethink cannot retrieve nil slices, so we force putting empty slices in Link
func formatLink(link *chainscript.Link) {
	if link.Meta.Tags == nil {
		link.Meta.Tags = []string{}
	}
	if link.Meta.Refs == nil {
		link.Meta.Refs = []*chainscript.LinkReference{}
	}
	if link.Signatures == nil {
		link.Signatures = []*chainscript.Signature{}
	}
}

// CreateLink implements github.com/stratumn/go-core/store.LinkWriter.CreateLink.
func (a *Store) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	if link.Meta.OutDegree >= 0 {
		return nil, types.WrapError(store.ErrOutDegreeNotSupported, errorcode.Unimplemented, store.Component, "could not create link")
	}

	prevLinkHash := link.Meta.GetPrevLinkHash()

	formatLink(link)

	linkHash, err := link.Hash()
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not hash link")
	}

	w := linkWrapper{
		ID:        linkHash,
		Content:   link,
		Priority:  link.Meta.Priority,
		UpdatedAt: time.Now().UTC(),
		MapID:     link.Meta.MapId,
		Tags:      link.Meta.Tags,
		Process:   link.Meta.Process.Name,
		Step:      link.Meta.Step,
	}

	if len(prevLinkHash) > 0 {
		w.PrevLinkHash = prevLinkHash
	}

	if err := a.links.Get(linkHash).Replace(&w).Exec(a.session); err != nil {
		return nil, types.WrapError(err, errorcode.Internal, store.Component, "could not create link")
	}

	linkEvent := store.NewSavedLinks(link)

	for _, c := range a.eventChans {
		c <- linkEvent
	}

	return linkHash, nil
}

// GetSegment implements github.com/stratumn/go-core/store.SegmentReader.GetSegment.
func (a *Store) GetSegment(ctx context.Context, linkHash chainscript.LinkHash) (*chainscript.Segment, error) {
	cur, err := a.links.Get(linkHash).Run(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get segment")
	}
	defer cur.Close()

	var w linkWrapper
	if err := cur.One(&w); err != nil {
		if err == rethink.ErrEmptyResult {
			return nil, nil
		}

		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get segment")
	}

	segment, err := w.Content.Segmentify()
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not segmentify")
	}

	if evidences, err := a.GetEvidences(ctx, segment.LinkHash()); evidences != nil && err == nil {
		segment.Meta.Evidences = evidences
	}

	return segment, nil
}

// FindSegments implements github.com/stratumn/go-core/store.SegmentReader.FindSegments.
func (a *Store) FindSegments(ctx context.Context, filter *store.SegmentFilter) (*types.PaginatedSegments, error) {
	if len(filter.Referencing) > 0 {
		return nil, types.WrapError(store.ErrReferencingNotSupported, errorcode.Unimplemented, store.Component, "could not find segments")
	}

	var prevLinkHash []byte
	q := a.links

	if filter.WithoutParent || len(filter.PrevLinkHash) > 0 {
		if len(filter.PrevLinkHash) > 0 {
			prevLinkHash = filter.PrevLinkHash
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
		ids := make([]interface{}, len(filter.LinkHashes))
		for i, v := range filter.LinkHashes {
			ids[i] = v
		}
		q = q.GetAll(ids...)
	}

	orderingFunction := rethink.Desc
	if filter.Reverse {
		orderingFunction = rethink.Asc
	}

	if mapIDs := filter.MapIDs; len(mapIDs) > 0 {
		ids := make([]interface{}, len(mapIDs))
		for i, v := range mapIDs {
			ids[i] = v
		}
		q = q.Filter(func(row rethink.Term) interface{} {
			return rethink.Expr(ids).Contains(row.Field("mapId"))
		})
	} else if prevLinkHash := filter.PrevLinkHash; len(prevLinkHash) > 0 || filter.WithoutParent {
		q = q.OrderBy(rethink.OrderByOpts{Index: "prevLinkHashOrder"})
	} else if linkHashes := filter.LinkHashes; len(linkHashes) > 0 {
		q = q.OrderBy(rethink.Asc("id"))
	} else {
		q = q.OrderBy(rethink.OrderByOpts{Index: orderingFunction("order")})
	}

	if process := filter.Process; len(process) > 0 {
		q = q.Filter(rethink.Row.Field("process").Eq(process))
	}

	if step := filter.Step; len(step) > 0 {
		q = q.Filter(rethink.Row.Field("step").Eq(step))
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
				"evidences": rethink.Branch(row.HasFields("right"), row.Field("right").Field("content"), types.EvidenceSlice{}),
			},
		}
	})

	cur, err := q.Skip(filter.Offset).Limit(filter.Limit).Run(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not find segments")
	}
	defer cur.Close()

	segments := &types.PaginatedSegments{
		Segments: make(types.SegmentSlice, 0, filter.Limit),
	}
	if err := cur.All(&segments.Segments); err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not find segments")
	}

	// Non optimal way to count all segments
	totalCountCur, err := q.Run(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not count segments")
	}
	defer totalCountCur.Close()

	segIt := chainscript.Segment{}
	for totalCountCur.Next(&segIt) {
		segments.TotalCount++
	}

	for _, s := range segments.Segments {
		err = s.SetLinkHash()
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not hash link")
		}
	}

	return segments, nil
}

// GetMapIDs implements github.com/stratumn/go-core/store.SegmentReader.GetMapIDs.
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
	} else {
		q = q.Between(rethink.MinVal, rethink.MaxVal, rethink.BetweenOpts{
			Index: "mapId",
		})
	}

	idFilter := false
	if filter.Prefix != "" {
		q = q.Filter(func(row rethink.Term) interface{} {
			return row.Field("mapId").Match(fmt.Sprintf("^%s.*", filter.Prefix))
		}).Field("mapId").Distinct()
		idFilter = true
	}
	if filter.Suffix != "" {
		q = q.Filter(func(row rethink.Term) interface{} {
			return row.Field("mapId").Match(fmt.Sprintf(".*%s$", filter.Suffix))
		}).Field("mapId").Distinct()
		idFilter = true
	}

	if !idFilter {
		if filter.Process != "" {
			q = q.OrderBy(rethink.OrderByOpts{Index: "processOrder"}).
				Distinct(rethink.DistinctOpts{Index: "processOrder"}).
				Map(func(row rethink.Term) interface{} {
					return row.AtIndex(1)
				})
		} else {
			q = q.OrderBy(rethink.OrderByOpts{Index: "mapId"}).
				Distinct(rethink.DistinctOpts{Index: "mapId"})
		}
	}

	cur, err := q.Skip(filter.Pagination.Offset).Limit(filter.Limit).Run(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get map ids")
	}
	defer cur.Close()

	mapIDs := []string{}
	if err = cur.All(&mapIDs); err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get map ids")
	}

	return mapIDs, nil
}

// AddEvidence implements github.com/stratumn/go-core/store.EvidenceWriter.AddEvidence.
func (a *Store) AddEvidence(ctx context.Context, linkHash chainscript.LinkHash, evidence *chainscript.Evidence) error {
	cur, err := a.evidences.Get(linkHash).Run(a.session)
	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not add evidence")
	}
	defer cur.Close()

	var ew evidencesWrapper
	if err := cur.One(&ew); err != nil {
		if err != rethink.ErrEmptyResult {
			return types.WrapError(err, errorcode.Unavailable, store.Component, "could not add evidence")
		}
	}

	currentEvidences := ew.Content
	if currentEvidences == nil {
		currentEvidences = types.EvidenceSlice{}
	}

	if err := currentEvidences.AddEvidence(evidence); err != nil {
		return err
	}

	w := evidencesWrapper{
		ID:        linkHash,
		Content:   currentEvidences,
		UpdatedAt: time.Now(),
	}
	if err := a.evidences.Get(linkHash).Replace(&w).Exec(a.session); err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not add evidence")
	}

	evidenceEvent := store.NewSavedEvidences()
	evidenceEvent.AddSavedEvidence(linkHash, evidence)

	for _, c := range a.eventChans {
		c <- evidenceEvent
	}
	return nil
}

// GetEvidences implements github.com/stratumn/go-core/store.EvidenceReader.GetEvidences.
func (a *Store) GetEvidences(ctx context.Context, linkHash chainscript.LinkHash) (types.EvidenceSlice, error) {
	cur, err := a.evidences.Get(linkHash).Run(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get evidences")
	}
	defer cur.Close()

	var ew evidencesWrapper
	if err := cur.One(&ew); err != nil {
		if err == rethink.ErrEmptyResult {
			return types.EvidenceSlice{}, nil
		}

		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get evidences")
	}
	return ew.Content, nil
}

// GetValue implements github.com/stratumn/go-core/store.KeyValueStore.GetValue.
func (a *Store) GetValue(ctx context.Context, key []byte) ([]byte, error) {
	cur, err := a.values.Get(key).Run(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get value")
	}
	defer cur.Close()

	var w valueWrapper
	if err := cur.One(&w); err != nil {
		if err == rethink.ErrEmptyResult {
			return nil, nil
		}

		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get value")
	}

	return w.Value, nil
}

// SetValue implements github.com/stratumn/go-core/store.KeyValueStore.SetValue.
func (a *Store) SetValue(ctx context.Context, key, value []byte) error {
	v := &valueWrapper{
		ID:    key,
		Value: value,
	}

	err := a.values.Get(key).Replace(&v).Exec(a.session)
	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not set value")
	}

	return nil
}

// DeleteValue implements github.com/stratumn/go-core/store.KeyValueStore.DeleteValue.
func (a *Store) DeleteValue(ctx context.Context, key []byte) ([]byte, error) {
	res, err := a.values.
		Get(key).
		Delete(rethink.DeleteOpts{ReturnChanges: true}).
		RunWrite(a.session)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not delete value")
	}
	if res.Deleted < 1 {
		return nil, nil
	}
	b, err := json.Marshal(res.Changes[0].OldValue)
	if err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "json.Marshal")
	}

	var w valueWrapper
	if err := json.Unmarshal(b, &w); err != nil {
		return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "json.Unmarshal")
	}

	return w.Value, nil
}

type rethinkBufferedBatch struct {
	*bufferedbatch.Batch
}

// CreateLink implements github.com/stratumn/go-core/store.LinkWriter.CreateLink.
func (b *rethinkBufferedBatch) CreateLink(ctx context.Context, link *chainscript.Link) (chainscript.LinkHash, error) {
	formatLink(link)
	return b.Batch.CreateLink(ctx, link)
}

// NewBatch implements github.com/stratumn/go-core/store.Adapter.NewBatch.
func (a *Store) NewBatch(ctx context.Context) (store.Batch, error) {
	bbBatch := bufferedbatch.NewBatch(ctx, a)
	if bbBatch == nil {
		return nil, types.NewError(errorcode.Internal, store.Component, "cannot create underlying batch")
	}

	return &rethinkBufferedBatch{Batch: bbBatch}, nil
}

// Create creates the database tables and indexes.
func (a *Store) Create() error {
	var err error
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}

	exists, err := a.Exists()
	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not create tables")
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

	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not create tables")
	}

	return nil
}

// Drop drops the database tables and indexes.
func (a *Store) Drop() error {
	var err error
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}

	exec(a.db.TableDrop("links"))
	exec(a.db.TableDrop("evidences"))
	exec(a.db.TableDrop("values"))

	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not drop tables")
	}

	return nil
}

// Clean removes all documents from the tables.
func (a *Store) Clean() error {
	var err error
	exec := func(term rethink.Term) {
		if err == nil {
			err = term.Exec(a.session)
		}
	}

	exec(a.links.Delete())
	exec(a.evidences.Delete())
	exec(a.values.Delete())

	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not clean tables")
	}

	return nil
}

// Exists returns whether the database tables exists.
func (a *Store) Exists() (bool, error) {
	cur, err := a.db.TableList().Run(a.session)
	if err != nil {
		return false, types.WrapError(err, errorcode.Unavailable, store.Component, "could not test tables")
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
