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
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"math/rand"

	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"
	"golang.org/x/crypto/ed25519"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/testutil"
)

// CreateLink creates a minimal link.
func CreateLink(process, mapID, prevLinkHash string, tags []interface{}, priority float64) *cs.Link {
	linkMeta := map[string]interface{}{
		"process":  process,
		"mapId":    mapID,
		"priority": priority,
		"random":   testutil.RandomString(12),
	}

	if prevLinkHash != "" {
		linkMeta["prevLinkHash"] = prevLinkHash
	}

	if tags != nil {
		linkMeta["tags"] = tags
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
	clone.Meta["mapId"] = testutil.RandomString(24)
	return clone
}

// RandomBranch appends a random link to a link.
func RandomBranch(parent *cs.Link) *cs.Link {
	linkHash, _ := parent.HashString()
	branch := CreateLink(testutil.RandomString(24), testutil.RandomString(24),
		linkHash, RandomTags(), rand.Float64())
	branch.Meta["mapId"] = parent.Meta["mapId"]
	return branch
}

// RandomTags creates between zero and four random tags.
func RandomTags() []interface{} {
	var tags []interface{}
	for i := 0; i < rand.Intn(5); i++ {
		tags = append(tags, testutil.RandomString(12))
	}
	return tags
}

// SignLink adds a signature to a link.
// The ed25519 signature algorithm is used.
func SignLink(l *cs.Link) *cs.Link {
	pub, priv, _ := ed25519.GenerateKey(crand.Reader)
	payloadPath := "[state, meta]"
	payload, _ := jmespath.Search(payloadPath, l)
	payloadBytes, _ := cj.Marshal(payload)
	sigBytes, _ := priv.Sign(crand.Reader, payloadBytes, crypto.Hash(0))
	sig := cs.Signature{
		Type:      "ed25519",
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		Signature: base64.StdEncoding.EncodeToString(sigBytes),
		Payload:   payloadPath,
	}
	l.Signatures = append(l.Signatures, &sig)
	return l
}

// SignLinkWithKey signs the link with the provided private key.
// The key must be an instance of ed25519.PrivateKey
func SignLinkWithKey(l *cs.Link, priv ed25519.PrivateKey) *cs.Link {
	pub := priv.Public().(ed25519.PublicKey)
	payloadPath := "[state, meta]"
	payload, _ := jmespath.Search(payloadPath, l)
	payloadBytes, _ := cj.Marshal(payload)
	sigBytes, _ := priv.Sign(crand.Reader, payloadBytes, crypto.Hash(0))
	sig := cs.Signature{
		Type:      "ed25519",
		PublicKey: base64.StdEncoding.EncodeToString(pub),
		Signature: base64.StdEncoding.EncodeToString(sigBytes),
		Payload:   payloadPath,
	}
	l.Signatures = append(l.Signatures, &sig)
	return l
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
