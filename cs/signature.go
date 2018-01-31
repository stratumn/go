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
