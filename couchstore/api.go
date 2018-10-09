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

package couchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/utils"
)

const (
	statusError           = 400
	statusDBExists        = 412
	statusDocumentMissing = 404
	statusDBMissing       = 404

	dbLink      = "pop_link"
	dbEvidences = "pop_evidences"
	dbValue     = "kv"

	objectTypeLink = "link"
	objectTypeMap  = "map"
)

// CouchResponseStatus contains couch specific response when querying the API.
type CouchResponseStatus struct {
	Ok         bool
	StatusCode int
	Error      string `json:"error,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (c *CouchResponseStatus) error() error {
	return errors.Errorf("Status code: %v, error: %v, reason: %v", c.StatusCode, c.Error, c.Reason)
}

// LinkWrapper wraps a link before saving it to a CouchDB document.
// Links omit empty values in their JSON representation, which can break
// CouchDB's filtering.
// So we make sure the LinkWrapper contains explicit default values for all
// fields we want to be able to sort on.
type LinkWrapper struct {
	Link         *chainscript.Link `json:"link"`
	Priority     float64           `json:"priority"`
	PrevLinkHash string            `json:"prevLinkHash"`
}

// WrapLink wraps a link.
func WrapLink(link *chainscript.Link) *LinkWrapper {
	wrapper := &LinkWrapper{
		Link:     link,
		Priority: link.Meta.Priority,
	}

	if len(link.PrevLinkHash()) > 0 {
		wrapper.PrevLinkHash = link.PrevLinkHash().String()
	}

	return wrapper
}

// Document is the object stored in CouchDB.
type Document struct {
	ID         string `json:"_id,omitempty"`
	Revision   string `json:"_rev,omitempty"`
	ObjectType string `json:"docType,omitempty"`

	// The following fields are used when querying couchdb for link documents.
	LinkWrapper *LinkWrapper `json:"linkWrapper,omitempty"`

	// The following fields are used when querying couchdb for evidences documents.
	Evidences types.EvidenceSlice `json:"evidences,omitempty"`

	// The following fields are used when querying couchdb for map documents.
	Process string `json:"process,omitempty"`

	// The following fields are used when querying couchdb for values stored via key/value.
	Value []byte `json:"value,omitempty"`
}

// BulkDocuments is used to bulk save documents to couchdb.
type BulkDocuments struct {
	Documents []*Document `json:"docs"`
	Atomic    bool        `json:"all_or_nothing,omitempty"`
}

// GetDatabases lists available databases.
func (c *CouchStore) GetDatabases() ([]string, error) {
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

// CreateDatabase creates a database.
func (c *CouchStore) CreateDatabase(dbName string) error {
	_, couchResponseStatus, err := c.put("/"+dbName, nil)
	if err != nil {
		return err
	}

	if !couchResponseStatus.Ok {
		if couchResponseStatus.StatusCode == statusDBExists {
			return nil
		}

		return couchResponseStatus.error()
	}

	return utils.Retry(func(attempt int) (bool, error) {
		path := fmt.Sprintf("/%s", dbName)
		_, couchResponseStatus, err := c.doHTTPRequest(http.MethodGet, path, nil)
		if err != nil || !couchResponseStatus.Ok {
			time.Sleep(200 * time.Millisecond)
			return true, err
		}
		return false, err
	}, 10)
}

// DeleteDatabase deletes a database.
func (c *CouchStore) DeleteDatabase(name string) error {
	_, couchResponseStatus, err := c.delete("/" + name)
	if err != nil {
		return err
	}

	if !couchResponseStatus.Ok {
		if couchResponseStatus.StatusCode != statusDBMissing {
			return couchResponseStatus.error()
		}
	}

	return nil
}

// CreateIndex creates an index.
func (c *CouchStore) CreateIndex(dbName string, indexName string, fields []string) error {
	path := fmt.Sprintf("/%s/_index", dbName)

	type createIndexDesc struct {
		Fields []string `json:"fields"`
	}
	type createIndexRequest struct {
		Name  string          `json:"name"`
		Index createIndexDesc `json:"index"`
	}

	payload := createIndexRequest{
		Name:  indexName,
		Index: createIndexDesc{Fields: fields},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, couchResponseStatus, err := c.post(path, payloadBytes)
	if err != nil {
		return err
	}

	if !couchResponseStatus.Ok {
		if couchResponseStatus.StatusCode != statusDBMissing {
			return couchResponseStatus.error()
		}
	}

	return nil
}

func (c *CouchStore) createLink(link *chainscript.Link) (chainscript.LinkHash, error) {
	linkHash, err := link.Hash()
	if err != nil {
		return nil, err
	}
	linkHashStr := linkHash.String()

	linkDoc := &Document{
		ObjectType:  objectTypeLink,
		LinkWrapper: WrapLink(link),
		ID:          linkHashStr,
	}

	currentLinkDoc, err := c.getDocument(dbLink, linkHashStr)
	if err != nil {
		return nil, err
	}
	if currentLinkDoc != nil {
		return nil, errors.Errorf("Link is immutable, %s already exists", linkHashStr)
	}

	docs := []*Document{
		linkDoc,
		{
			ObjectType: objectTypeMap,
			ID:         linkDoc.LinkWrapper.Link.Meta.MapId,
			Process:    linkDoc.LinkWrapper.Link.Meta.Process.Name,
		},
	}

	return linkHash, c.saveDocuments(dbLink, docs)
}

func (c *CouchStore) addEvidence(linkHash string, evidence *chainscript.Evidence) error {
	currentDoc, err := c.getDocument(dbEvidences, linkHash)
	if err != nil {
		return err
	}
	if currentDoc == nil {
		currentDoc = &Document{
			ID: linkHash,
		}
	}
	if currentDoc.Evidences == nil {
		currentDoc.Evidences = types.EvidenceSlice{}
	}

	if err := currentDoc.Evidences.AddEvidence(evidence); err != nil {
		return err
	}

	return c.saveDocument(dbEvidences, linkHash, *currentDoc)
}

func (c *CouchStore) segmentify(ctx context.Context, link *chainscript.Link) *chainscript.Segment {
	segment, err := link.Segmentify()
	if err != nil {
		panic(err)
	}

	if evidences, err := c.GetEvidences(ctx, segment.LinkHash()); evidences != nil && err == nil {
		segment.Meta.Evidences = evidences
	}

	return segment
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
	if !couchResponseStatus.Ok {
		return couchResponseStatus.error()
	}

	return nil
}

func (c *CouchStore) saveDocuments(dbName string, docs []*Document) error {
	bulkDocuments := BulkDocuments{
		Documents: docs,
	}

	path := fmt.Sprintf("/%v/_bulk_docs", dbName)

	docsBytes, err := json.Marshal(bulkDocuments)
	if err != nil {
		return err
	}

	_, _, err = c.post(path, docsBytes)
	return err
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

	if !couchResponseStatus.Ok {
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

	if !couchResponseStatus.Ok {
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
