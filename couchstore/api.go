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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/utils"
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

func (c *CouchResponseStatus) error() error {
	return errors.Errorf("Status code: %v, error: %v, reason: %v", c.StatusCode, c.Error, c.Reason)
}

// Document is the type used in couchdb
type Document struct {
	ID         string `json:"_id,omitempty"`
	Revision   string `json:"_rev,omitempty"`
	ObjectType string `json:"docType,omitempty"`

	// The following fields are used when querying couchdb for segment documents.
	Segment *cs.Segment `json:"segment,omitempty"`

	// The following fields are used when querying couchdb for map documents
	Process string `json:"process,omitempty"`

	// The following fields are used when querying couchdb for values stored via key/value.
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

func (c *CouchStore) createDatabase(dbName string) error {
	_, couchResponseStatus, err := c.put("/"+dbName, nil)
	if err != nil {
		return err
	}

	if couchResponseStatus.Ok == false {
		if couchResponseStatus.StatusCode != statusDBExists {
			return couchResponseStatus.error()
		}
	}

	return utils.Retry(func(attempt int) (bool, error) {
		path := fmt.Sprintf("/%s", dbName)
		_, couchResponseStatus, err := c.doHTTPRequest(http.MethodGet, path, nil)
		if err != nil || couchResponseStatus.Ok == false {
			time.Sleep(200 * time.Millisecond)
			return true, err
		}
		return false, err
	}, 10)
}

func (c *CouchStore) deleteDatabase(name string) error {
	_, couchResponseStatus, err := c.delete("/" + name)
	if err != nil {
		return err
	}

	if couchResponseStatus.Ok == false {
		if couchResponseStatus.StatusCode != statusDBMissing {
			return couchResponseStatus.error()
		}
	}

	return nil
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
		return couchResponseStatus.error()
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
		return nil, couchResponseStatus.error()
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
	return c.doHTTPRequest(http.MethodGet, path, nil)
}

func (c *CouchStore) post(path string, data []byte) ([]byte, *CouchResponseStatus, error) {
	resp, err := http.Post(c.config.Address+path, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, nil, err
	}

	return getCouchResponseStatus(resp)
}

func (c *CouchStore) put(path string, data []byte) ([]byte, *CouchResponseStatus, error) {
	return c.doHTTPRequest(http.MethodPut, path, data)
}

func (c *CouchStore) delete(path string) ([]byte, *CouchResponseStatus, error) {
	return c.doHTTPRequest(http.MethodDelete, path, nil)
}

func (c *CouchStore) doHTTPRequest(method string, path string, data []byte) ([]byte, *CouchResponseStatus, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, c.config.Address+path, bytes.NewBuffer(data))
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
