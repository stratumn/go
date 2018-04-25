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

package cstesting

import (
	"crypto"

	"github.com/stratumn/go-indigocore/cs"
)

// LinkBuilder allows building links easily in tests.
type LinkBuilder struct {
	Link *cs.Link
}

// NewLinkBuilder creates a new LinkBuilder.
func NewLinkBuilder() *LinkBuilder {
	return &LinkBuilder{
		Link: RandomLink(),
	}
}

// Invalid makes the link invalid by erasing the mapID.
func (lb *LinkBuilder) Invalid() *LinkBuilder {
	// a link without a mapID is invalid
	lb.Link.Meta.MapID = ""
	return lb
}

// WithState fills the link's state.
func (lb *LinkBuilder) WithState(state map[string]interface{}) *LinkBuilder {
	lb.Link.State = state
	return lb
}

// WithPrevLinkHash fills the link's prevLinkHash.
func (lb *LinkBuilder) WithPrevLinkHash(prevLinkHash string) *LinkBuilder {
	lb.Link.Meta.PrevLinkHash = prevLinkHash
	return lb
}

// WithParent fills the link's prevLinkHash with the given parent's hash.
func (lb *LinkBuilder) WithParent(link *cs.Link) *LinkBuilder {
	linkHash, _ := link.HashString()
	lb.Link.Meta.PrevLinkHash = linkHash
	return lb
}

// WithoutParent removes the link's parent (prevLinkHash).
func (lb *LinkBuilder) WithoutParent() *LinkBuilder {
	lb.Link.Meta.PrevLinkHash = ""
	return lb
}

// WithTag adds a tag to the link.
func (lb *LinkBuilder) WithTag(tag string) *LinkBuilder {
	lb.Link.Meta.Tags = append(lb.Link.Meta.Tags, tag)
	return lb
}

// WithTags replaces the link's tags.
func (lb *LinkBuilder) WithTags(tags ...string) *LinkBuilder {
	lb.Link.Meta.Tags = tags
	return lb
}

// WithMapID fills the link's mapID.
func (lb *LinkBuilder) WithMapID(mapID string) *LinkBuilder {
	lb.Link.Meta.MapID = mapID
	return lb
}

// WithProcess fills the link's process.
func (lb *LinkBuilder) WithProcess(process string) *LinkBuilder {
	lb.Link.Meta.Process = process
	return lb
}

// WithType fills the link's type.
func (lb *LinkBuilder) WithType(linkType string) *LinkBuilder {
	lb.Link.Meta.Type = linkType
	return lb
}

// WithMetadata adds an entry in the Meta.Data map.
func (lb *LinkBuilder) WithMetadata(key string, value interface{}) *LinkBuilder {
	if lb.Link.Meta.Data == nil {
		lb.Link.Meta.Data = make(map[string]interface{})
	}

	lb.Link.Meta.Data[key] = value
	return lb
}

// WithRef adds a reference to the link.
func (lb *LinkBuilder) WithRef(link *cs.Link) *LinkBuilder {
	refHash, _ := link.HashString()
	lb.Link.Meta.Refs = append(lb.Link.Meta.Refs, cs.SegmentReference{
		LinkHash: refHash,
		Process:  link.Meta.Process,
	})
	return lb
}

// Branch uses the provided link as its parent and copies its mapID and process.
func (lb *LinkBuilder) Branch(parent *cs.Link) *LinkBuilder {
	lh, _ := parent.HashString()
	lb.Link.Meta.PrevLinkHash = lh
	lb.Link.Meta.MapID = parent.Meta.MapID
	lb.Link.Meta.Process = parent.Meta.Process
	return lb
}

// Sign signs the link with a random signature.
func (lb *LinkBuilder) Sign() *LinkBuilder {
	lb.Link.Signatures = append(lb.Link.Signatures, RandomSignature(lb.Link))
	return lb
}

// SignWithKey signs the link with the provided private key.
func (lb *LinkBuilder) SignWithKey(priv crypto.PrivateKey) *LinkBuilder {
	lb.Link.Signatures = append(lb.Link.Signatures, SignatureWithKey(lb.Link, priv))
	return lb
}

// From assigns a clone of the provided link to its internal link
func (lb *LinkBuilder) From(l *cs.Link) *LinkBuilder {
	lb.Link = Clone(l)
	return lb
}

// Build returns the underlying link.
func (lb *LinkBuilder) Build() *cs.Link {
	return lb.Link
}
