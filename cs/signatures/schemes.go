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
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

const (
	// Ed25519 is the EdDSA signature scheme using SHA-512/256 and Curve25519.
	Ed25519 = "ED25519"

	// ECDSA256 is the ecdsa scheme using a P-256 ellipitc curve
	ECDSA256 = "ECDSA256"

	// RSA is the digital signature scheme as defined in PKCS#1 (RSASSA-PKCS1-v1_5)
	RSA = "RSA"
)

var (
	// SupportedSignatureTypes is a list of the supported signature types.
	supportedSignatureTypes = []string{Ed25519, ECDSA256, RSA}

	// ErrInvalidSignature is returned when the signature verification failed.
	ErrInvalidSignature = errors.New("signature verification failed")

	// ErrUnsupportedSignatureType is returned when the signature type is not supported.
	ErrUnsupportedSignatureType = errors.Errorf("signature type must be one of %v", supportedSignatureTypes)
)

// MatchScheme returns the signature scheme if it is currently supported and an error otherwise
func MatchScheme(sigType string) (string, error) {
	for _, supportedSig := range supportedSignatureTypes {
		if strings.EqualFold(sigType, supportedSig) {
			return supportedSig, nil
		}
	}
	return "", errors.Wrapf(ErrUnsupportedSignatureType, "Unhandled signature scheme [%s]", sigType)
}

// Verify checks that the signature scheme is supported and runs the verification for this scheme.
// It returns an error if the verification failed and nil otherwise.
// _______________________________________
// Public keys must be encoded as follows:
// - Ed25519: no encoding necessary, public key length should be 32 bytes.
// - ECDSA256: point (X, Y) of the public key must be encoded into the form specified in section 4.3.6 of ANSI X9.62 ("Point-to-Octet-String Conversion").
// - RSA: public key (modulus, public exponent) must be ASN-1 encoded.
// ______________________________________
// Signatures must be encoded as follows:
// - Ed25519: no encoding necessary, signature length should be 64 bytes.
// - ECDSA256: ecdsa signatures (R, S *big.Int) must be encoded to ASN-1.
// - RSA: signatures must be computed using PKCS1-v1_5. What should be signed is the SHA512 sum of the document, not the document itself.
func Verify(signatureType string, key, signature, document []byte) error {
	scheme, err := MatchScheme(signatureType)
	if err != nil {
		return err
	}

	switch scheme {
	case Ed25519:
		publicKey := ed25519.PublicKey(key)
		if len(publicKey) != ed25519.PublicKeySize {
			return errors.Errorf("Ed25519 public key length must be %d, got %d", ed25519.PublicKeySize, len(publicKey))
		}
		if !ed25519.Verify(publicKey, document, signature) {
			return ErrInvalidSignature
		}
	case ECDSA256:
		publicKey, err := newECDSAPublicKey(elliptic.P256(), key)
		if err != nil {
			return err
		}
		ecdsaSig, err := newECDSASignature(signature)
		if err != nil {
			return errors.Wrap(err, "Could not parse ECDSA signature")
		}
		if !ecdsa.Verify(publicKey, document, ecdsaSig.R, ecdsaSig.S) {
			return ErrInvalidSignature
		}
	case RSA:
		publicKey, err := newRSAPublicKey(key)
		if err != nil {
			return err
		}
		hashedDocument := sha512.Sum512(document)
		if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA512, hashedDocument[:], signature); err != nil {
			return errors.Wrap(err, ErrInvalidSignature.Error())
		}
	}
	return nil
}

// Sign signs a message with the private key.
// It returns the public key and the signature if the message was signed correctly and an error otherwise.
// _______________________________________
// Private keys must be encoded as follows:
// - Ed25519: no encoding necessary, private key length should be 64 bytes.
func Sign(signatureType string, privateKey, msg []byte) (publicKey string, signature string, err error) {
	scheme, err := MatchScheme(signatureType)
	if err != nil {
		return "", "", err
	}

	switch scheme {
	case Ed25519:
		if len(privateKey) != ed25519.PrivateKeySize {
			return "", "", errors.Errorf("%s private key length must be %d, got %d", Ed25519, ed25519.PrivateKeySize, len(privateKey))
		}
		pk := ed25519.PrivateKey(privateKey)

		publicKey = base64.StdEncoding.EncodeToString(pk.Public().(ed25519.PublicKey))
		signatureBytes, err := pk.Sign(rand.Reader, msg, crypto.Hash(0))
		if err != nil {
			return "", "", err
		}
		signature = base64.StdEncoding.EncodeToString(signatureBytes)
	case ECDSA256:
		return "", "", ErrUnsupportedSignatureType
	case RSA:
		return "", "", ErrUnsupportedSignatureType
	}

	return publicKey, signature, nil
}
