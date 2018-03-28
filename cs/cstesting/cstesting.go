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

// Package cstesting defines helpers to test Chainscripts.
package cstesting

import (
	"crypto"
	"encoding/json"
	"math/rand"

	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"

	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-crypto/signatures"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/testutil"
)

// CreateLink creates a minimal link.
func CreateLink(process, mapID, prevLinkHash string, tags []string, priority float64) *cs.Link {
	linkMeta := cs.LinkMeta{
		Process:      process,
		MapID:        mapID,
		PrevLinkHash: prevLinkHash,
		Tags:         tags,
		Priority:     priority,
		Action:       testutil.RandomString(24),
		Type:         testutil.RandomString(24),
		Inputs:       RandomInterfaces(),
		Refs:         []cs.SegmentReference{},
		Data: map[string]interface{}{
			"random": testutil.RandomString(12),
		},
	}

	link := &cs.Link{
		State: map[string]interface{}{
			"random": testutil.RandomString(12),
		},
		Meta:       linkMeta,
		Signatures: cs.Signatures{},
	}

	return link
}

// RandomLink creates a random link.
func RandomLink() *cs.Link {
	return CreateLink(testutil.RandomString(24), testutil.RandomString(24),
		testutil.RandomHash().String(), RandomTags(), rand.Float64())
}

// RandomSegment creates a random segment.
func RandomSegment() *cs.Segment {
	return RandomLink().Segmentify()
}

// RandomLinkWithProcess creates a random link in a specific process.
func RandomLinkWithProcess(process string) *cs.Link {
	return CreateLink(process, testutil.RandomString(24),
		testutil.RandomHash().String(), RandomTags(), rand.Float64())
}

// InvalidLinkWithProcess creates a random invalid link.
func InvalidLinkWithProcess(process string) *cs.Link {
	// A link with no MapId is invalid
	return CreateLink(process, "",
		testutil.RandomHash().String(), RandomTags(), rand.Float64())
}

// RandomEvidence creates a random evidence.
func RandomEvidence() *cs.Evidence {
	return &cs.Evidence{
		Provider: testutil.RandomString(12),
		Backend:  "generic",
	}
}

// ChangeState clones a link and randomly changes its state.
func ChangeState(l *cs.Link) *cs.Link {
	clone := Clone(l)
	clone.State["random"] = testutil.RandomString(12)
	return clone
}

// ChangeMapID clones a link and randomly changes its map ID.
func ChangeMapID(l *cs.Link) *cs.Link {
	clone := Clone(l)
	clone.Meta.MapID = testutil.RandomString(24)
	return clone
}

// RandomBranch appends a random link to a link.
func RandomBranch(parent *cs.Link) *cs.Link {
	linkHash, _ := parent.HashString()
	branch := CreateLink(testutil.RandomString(24), testutil.RandomString(24),
		linkHash, RandomTags(), rand.Float64())
	branch.Meta.MapID = parent.Meta.MapID
	return branch
}

// RandomTags creates between zero and four random tags.
func RandomTags() []string {
	var tags []string
	for i := 0; i < rand.Intn(5); i++ {
		tags = append(tags, testutil.RandomString(12))
	}
	return tags
}

// RandomInterfaces creates between zero and four random values of type string/float/bool.
// int type is not generated because of assertion failure on float/int interpretation
func RandomInterfaces() []interface{} {
	var ret []interface{}
	for i := 0; i < rand.Intn(5); i++ {
		switch rand.Intn(3) {
		case 0:
			ret = append(ret, testutil.RandomString(12))
		case 1:
			ret = append(ret, rand.Float64())
		case 2:
			ret = append(ret, rand.Int() < 42)
		}
	}
	return ret
}

// SignLink adds a signature to a link.
// The ed25519 signature algorithm is used.
func SignLink(l *cs.Link) *cs.Link {
	l.Signatures = append(l.Signatures, RandomSignature(l))
	return l
}

// SignLinkWithKey signs the link with the provided private key.
// The key must be an instance of ed25519.PrivateKey
func SignLinkWithKey(l *cs.Link, priv crypto.PrivateKey) *cs.Link {
	l.Signatures = append(l.Signatures, SignatureWithKey(l, priv))
	return l
}

// RandomSignature returns an arbitrary signature from a generated key pair
func RandomSignature(l *cs.Link) *cs.Signature {
	_, priv, _ := keys.GenerateKey(keys.ED25519)
	payloadPath := "[state, meta]"
	payload, _ := jmespath.Search(payloadPath, l)
	payloadBytes, _ := cj.Marshal(payload)
	sig, _ := signatures.Sign(priv, payloadBytes)
	return &cs.Signature{
		Type:      sig.AI,
		PublicKey: string(sig.PublicKey),
		Signature: string(sig.Signature),
		Payload:   payloadPath,
	}
}

// SignatureWithKey returns a signature of a link using the provided private key
func SignatureWithKey(l *cs.Link, priv crypto.PrivateKey) *cs.Signature {
	privPEM, _ := keys.EncodeSecretkey(priv)
	payloadPath := "[state, meta]"
	payload, _ := jmespath.Search(payloadPath, l)
	payloadBytes, _ := cj.Marshal(payload)
	sig, _ := signatures.Sign(privPEM, payloadBytes)
	return &cs.Signature{
		Type:      sig.AI,
		PublicKey: string(sig.PublicKey),
		Signature: string(sig.Signature),
		Payload:   payloadPath,
	}
}

// Clone clones a link.
func Clone(l *cs.Link) *cs.Link {
	var clone cs.Link

	js, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(js, &clone); err != nil {
		panic(err)
	}

	return &clone
}
