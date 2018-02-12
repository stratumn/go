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
	"crypto/rsa"
	"encoding/asn1"

	"github.com/pkg/errors"
)

const (
	// RSAKeySize is the length of the rsa key in bits
	RSAKeySize = 2048
)

var (
	// ErrInvalidRSAPublicKey is returned when the unmarshalling of an rsa public key failed
	ErrInvalidRSAPublicKey = errors.New("Could not parse RSA public key, wrong (N, E) parameters")
)

func newRSAPublicKey(key []byte) (*rsa.PublicKey, error) {
	var publicKey rsa.PublicKey
	_, err := asn1.Unmarshal(key, &publicKey)
	if err != nil {
		return nil, errors.Wrap(ErrInvalidRSAPublicKey, "Error while decoding public key")
	}
	return &publicKey, nil
}
