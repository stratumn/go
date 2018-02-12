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

package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"math/big"

	"github.com/pkg/errors"
)

var (
	// ErrInvalidECDSAPublicKey is returned when the unmarshalling of an ecdsa public key failed
	ErrInvalidECDSAPublicKey = errors.New("Could not parse ECDSA public key, wrong (X, Y) parameters")
)

// ecdsaSignature is used to encode/decode ECDSA signatures to/from ASN-1.
// It is used the same way as in the "crypto/ecdsa" package (which does not export this type)
type ecdsaSignature struct {
	R, S *big.Int
}

// newECDSASignature serializes a bytes-formatted signature into a pair of bigInts
func newECDSASignature(bytesSignature []byte) (*ecdsaSignature, error) {
	var sig ecdsaSignature
	if _, err := asn1.Unmarshal(bytesSignature, &sig); err != nil {
		return nil, err
	}
	return &sig, nil
}

func newECDSAPublicKey(curve elliptic.Curve, key []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(curve, key)
	if x == nil {
		return nil, ErrInvalidECDSAPublicKey
	}
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}
