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
	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"

	"github.com/stratumn/go-crypto/signatures"
)

const (
	// ErrBadJMESPATHQuery is returned when the JMESPATH engine fails to execute the query
	ErrBadJMESPATHQuery = "failed to execute JMESPATH query"

	// ErrEmptyJMESPATHResult is returned when the JMESPATH query does not match any part of the link
	ErrEmptyJMESPATHResult = "JMESPATH query does not match any link data"
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

// NewSignature creates a new signature for a link.
// Only the data matching the JMESPATH query will be signed
func NewSignature(payloadPath string, privateKey []byte, l *Link) (*Signature, error) {
	payload, err := jmespath.Search(payloadPath, l)
	if err != nil {
		return nil, errors.Wrap(err, ErrBadJMESPATHQuery)
	}
	if payload == nil {
		return nil, errors.New(ErrEmptyJMESPATHResult)
	}

	payloadBytes, err := cj.Marshal(payload)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signature, err := signatures.Sign(privateKey, payloadBytes)
	if err != nil {
		return nil, err
	}

	return &Signature{
		Type:      signature.AI,
		PublicKey: string(signature.PublicKey),
		Signature: string(signature.Signature),
		Payload:   payloadPath,
	}, nil
}

// Verify takes a link as input, computes the signed part using the signature payload
// and runs the signature verification depending on its type.
func (s Signature) Verify(l *Link) error {
	payload, err := jmespath.Search(s.Payload, l)
	if err != nil {
		return errors.Wrap(err, ErrBadJMESPATHQuery)
	}
	if payload == nil {
		return errors.New(ErrEmptyJMESPATHResult)
	}

	payloadBytes, err := cj.Marshal(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := signatures.Verify(&signatures.Signature{
		AI:        s.Type,
		PublicKey: []byte(s.PublicKey),
		Message:   payloadBytes,
		Signature: []byte(s.Signature),
	}); err != nil {
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
