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

package couchstore

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/types"
)

const (
	statusError           = 400
	statusDBExists        = 412
	statusDocumentMissing = 404
	statusDBMissing       = 404

	dbSegment = "pop"
	dbMap     = "pop"
	dbValue   = "kv"

	objectTypeSegment = "segment"
	objectTypeMap     = "map"
)

// CouchResponseStatus contains couch specific response when querying the API.
type CouchResponseStatus struct {
	Ok         bool
	StatusCode int
	Error      string `json:"error;omitempty"`
	Reason     string `json:"reason;omitempty"`
}

// Document is the type used in couchdb
type Document struct {
	ID         string `json:"_id,omitempty"`
	Revision   string `json:"_rev,omitempty"`
	ObjectType string `json:"docType,omitempty"`

	// Segment specific
	Segment *cs.Segment `json:"segment,omitempty"`

	// MapID specific
	Process string `json:"process,omitempty"`

	// Value specific
	Value []byte `json:"value,omitempty"`
}

// BulkDocuments is used to bulk save documents to couchdb.
type BulkDocuments struct {
	Documents []*Document `json:"docs"`
	Atomic    bool        `json:"all_or_nothing,omitempty"`
}

func (c *CouchStore) getDatabases() ([]string, error) {
	body, _, err := c.get("/_all_dbs")
	if err != nil {
		return nil, err
	}

	databases := &[]string{}
	if err := json.Unmarshal(body, databases); err != nil {
		return nil, err
	}
	return *databases, nil
}

func (c *CouchStore) createDatabase(name string) error {
	_, couchResponseStatus, err := c.put("/"+name, nil)
	if err != nil {
		return err
	}

	if couchResponseStatus.Ok == false {
		if couchResponseStatus.StatusCode != statusDBExists {
			return errors.New(couchResponseStatus.Reason)
		}
	}

	return nil
}

func (c *CouchStore) deleteDatabase(name string) error {
	_, couchResponseStatus, err := c.delete("/" + name)
	if err != nil {
		return err
	}

	if couchResponseStatus.Ok == false {
		if couchResponseStatus.StatusCode != statusDBExists {
			return errors.New(couchResponseStatus.Error)
		}
	}

	return nil
}

func (c *CouchStore) getSegmentDoc(linkHash *types.Bytes32) (*Document, error) {
	return c.getDocument(dbSegment, linkHash.String())
}

func (c *CouchStore) getValueDoc(key []byte) (*Document, error) {
	return c.getDocument(dbValue, hex.EncodeToString(key))
}

func (c *CouchStore) saveSegment(segment *cs.Segment) error {
	segmentDoc := &Document{
		ObjectType: objectTypeSegment,
		Segment:    segment,
		ID:         segment.GetLinkHashString(),
	}

	currentSegmentDoc, err := c.getDocument(dbSegment, segment.GetLinkHashString())
	if err != nil {
		return err
	}
	if currentSegmentDoc != nil {
		if segment, err = currentSegmentDoc.Segment.MergeMeta(segment); err != nil {
			return err
		}
		segmentDoc = currentSegmentDoc
	}

	docs := []*Document{
		segmentDoc,
		&Document{
			ObjectType: objectTypeMap,
			ID:         segmentDoc.Segment.Link.GetMapID(),
			Process:    segmentDoc.Segment.Link.GetProcess(),
		},
	}
	bulkDocuments := BulkDocuments{
		Documents: docs,
	}

	path := fmt.Sprintf("/%v/_bulk_docs", dbSegment)

	docsBytes, err := json.Marshal(bulkDocuments)
	if err != nil {
		return err
	}

	_, _, err = c.post(path, docsBytes)
	return err
}

func (c *CouchStore) saveDocument(dbName string, key string, doc Document) error {
	path := fmt.Sprintf("/%v/%v", dbName, key)
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	_, couchResponseStatus, err := c.put(path, docBytes)
	if err != nil {
		return err
	}
	if couchResponseStatus.Ok == false {
		return errors.New(couchResponseStatus.Reason)
	}

	return nil
}

func (c *CouchStore) getDocument(dbName string, key string) (*Document, error) {
	doc := &Document{}
	path := fmt.Sprintf("/%v/%v", dbName, key)
	docBytes, couchResponseStatus, err := c.get(path)
	if err != nil {
		return nil, err
	}

	if couchResponseStatus.StatusCode == statusDocumentMissing {
		return nil, nil
	}

	if couchResponseStatus.Ok == false {
		return nil, errors.New(couchResponseStatus.Reason)
	}

	if err := json.Unmarshal(docBytes, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (c *CouchStore) deleteDocument(dbName string, key string) (*Document, error) {
	doc, err := c.getDocument(dbName, key)
	if err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, nil
	}

	path := fmt.Sprintf("/%v/%v?rev=%v", dbName, key, doc.Revision)
	_, couchResponseStatus, err := c.delete(path)
	if err != nil {
		return nil, err
	}

	if couchResponseStatus.Ok == false {
		return nil, errors.New(couchResponseStatus.Reason)
	}

	return doc, nil
}

func (c *CouchStore) get(path string) ([]byte, *CouchResponseStatus, error) {
	resp, err := http.Get(c.config.Address + path)
	if err != nil {
		return nil, nil, err
	}

	return getCouchResponseStatus(resp)
}

func (c *CouchStore) post(path string, data []byte) ([]byte, *CouchResponseStatus, error) {
	resp, err := http.Post(c.config.Address+path, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, nil, err
	}

	return getCouchResponseStatus(resp)
}

func (c *CouchStore) put(path string, data []byte) ([]byte, *CouchResponseStatus, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, c.config.Address+path, bytes.NewBuffer(data))
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return getCouchResponseStatus(resp)
}

func (c *CouchStore) delete(path string) ([]byte, *CouchResponseStatus, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, c.config.Address+path, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return getCouchResponseStatus(resp)
}

func getCouchResponseStatus(resp *http.Response) ([]byte, *CouchResponseStatus, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	couchResponseStatus := &CouchResponseStatus{}
	if resp.StatusCode >= statusError {
		if err := json.Unmarshal(body, couchResponseStatus); err != nil {
			return nil, nil, err
		}
		couchResponseStatus.Ok = false
	} else {
		couchResponseStatus.Ok = true
	}
	couchResponseStatus.StatusCode = resp.StatusCode

	return body, couchResponseStatus, nil
}
