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

// Package cs defines types to work with Chainscripts.
package cs

import (
	"encoding/base64"

	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"

	cj "github.com/gibson042/canonicaljson-go"

	signatures "github.com/stratumn/go-indigocore/cs/signatures"
)

// Signature contains a user-provided signature of a certain part of the link.
type Signature struct {
	// Type of the signature (eg: "EdDSA")
	Type string `json:"type"`

	// PublicKey is the base64 encoded public key that signed the payload
	PublicKey string `json:"publicKey"`

	// Signature is the base64 encoded string containg the signature bytes
	Signature string `json:"signature"`

	// Payload describes what has been signed, It is expressed using the JMESPATH query language.
	Payload string `json:"payload"`
}

// Verify takes a link as input, computes the signed part using the signature payload
// and runs the signature verification depending on its type.
func (s Signature) Verify(l *Link) error {
	keyBytes, _ := base64.StdEncoding.DecodeString(s.PublicKey)
	sigBytes, _ := base64.StdEncoding.DecodeString(s.Signature)

	payload, err := jmespath.Search(s.Payload, l)
	if err != nil {
		return errors.Wrap(err, "failed to execute JMESPATH query")
	}
	if payload == nil {
		return errors.New("JMESPATH query does not match any link data")
	}

	payloadBytes, err := cj.Marshal(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := signatures.Verify(s.Type, keyBytes, sigBytes, payloadBytes); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Signatures is a slice of Signature
type Signatures []*Signature

// Get returns the signature matching the provided public key
func (s *Signatures) Get(publicKey string) *Signature {
	for _, sig := range *s {
		if sig.PublicKey == publicKey {
			return sig
		}
	}
	return nil
}
