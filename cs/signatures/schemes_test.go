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
	"encoding/asn1"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

func TestMatchScheme(t *testing.T) {

	t.Run("Known schemes", func(t *testing.T) {
		schemes := map[string]string{"ed25519": Ed25519, "ecdsa256": ECDSA256, "rsa": RSA}
		for s, known := range schemes {
			got, err := MatchScheme(s)
			assert.NoError(t, err)
			assert.Equal(t, known, got)
		}
	})

	t.Run("Unknown scheme", func(t *testing.T) {
		scheme := "test"
		got, err := MatchScheme(scheme)
		assert.EqualError(t, err, "Unhandled signature scheme [test]: signature type must be one of [ED25519 ECDSA256 RSA]")
		assert.Empty(t, got)
	})
}
func TestVerify(t *testing.T) {

	document := []byte("test")

	t.Run("Ed25519", func(t *testing.T) {

		pub, priv, _ := ed25519.GenerateKey(rand.Reader)
		sig := ed25519.Sign(priv, document)

		t.Run("Valid signature", func(t *testing.T) {
			err := Verify(Ed25519, pub, sig, document)
			assert.NoError(t, err)
		})

		t.Run("Invalid signature", func(t *testing.T) {
			err := Verify(Ed25519, pub, []byte("test"), document)
			assert.EqualError(t, err, "signature verification failed")
		})

		t.Run("Bad public key length", func(t *testing.T) {
			err := Verify(Ed25519, []byte("test"), sig, document)
			assert.EqualError(t, err, "Ed25519 public key length must be 32, got 4")
		})
	})

	t.Run("ECDSA256", func(t *testing.T) {
		priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		r, s, err := ecdsa.Sign(rand.Reader, priv, document)
		require.NoError(t, err)

		// priv.X and priv.Y correspond to the public key. It is encoded using elliptic.Marshal.
		pub := elliptic.Marshal(elliptic.P256(), priv.X, priv.Y)

		// sig must me ASN-1 encoded
		sig, err := asn1.Marshal(ecdsaSignature{r, s})
		require.NoError(t, err)

		t.Run("Valid signature", func(t *testing.T) {
			err := Verify(ECDSA256, pub, sig, document)
			assert.NoError(t, err)
		})

		t.Run("Invalid signature", func(t *testing.T) {
			r := big.NewInt(0)
			s := big.NewInt(0)
			sig, _ := asn1.Marshal(ecdsaSignature{r, s})
			err = Verify(ECDSA256, pub, sig, document)
			assert.EqualError(t, err, "signature verification failed")
		})

		t.Run("Bad public key format", func(t *testing.T) {
			err := Verify(ECDSA256, []byte("test"), sig, document)
			assert.EqualError(t, err, "Could not parse ECDSA public key, wrong (X, Y) parameters")
		})

		t.Run("Bad signature format", func(t *testing.T) {
			err := Verify(ECDSA256, pub, []byte("test"), document)
			assert.Error(t, err, "Ed25519 public key length must be 32, got 4")
		})
	})

	t.Run("RSA", func(t *testing.T) {
		priv, _ := rsa.GenerateKey(rand.Reader, RSAKeySize)

		// sign the SHA512 hash of the document
		hashed := sha512.Sum512(document)
		sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA512, hashed[:])
		require.NoError(t, err)

		// RSA public key must me ASN-1 encoded
		pub, err := asn1.Marshal(priv.PublicKey)
		require.NoError(t, err)

		t.Run("Valid signature", func(t *testing.T) {
			err := Verify(RSA, pub, sig, document)
			assert.NoError(t, err)
		})

		t.Run("Invalid signature", func(t *testing.T) {
			sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA512, []byte(""))
			err = Verify(RSA, pub, sig, document)
			assert.EqualError(t, err, "signature verification failed: crypto/rsa: verification error")
		})

		t.Run("Bad public key format", func(t *testing.T) {
			err := Verify(RSA, []byte("test"), sig, document)
			assert.EqualError(t, err, "Error while decoding public key: Could not parse RSA public key, wrong (N, E) parameters")
		})
	})
}

func TestSign(t *testing.T) {

	document := []byte("test")
	_, priv, _ := ed25519.GenerateKey(rand.Reader)

	t.Run("Ed25519", func(t *testing.T) {

		t.Run("Valid signature", func(t *testing.T) {
			pub, sig, err := Sign(Ed25519, priv[:], document)
			assert.NoError(t, err)
			decodedSig, _ := base64.StdEncoding.DecodeString(sig)
			decodedPK, _ := base64.StdEncoding.DecodeString(pub)
			verified := ed25519.Verify(ed25519.PublicKey(decodedPK), document, decodedSig)
			assert.True(t, verified, "Ed25519 signature is invalid")
		})

		t.Run("Bad private key", func(t *testing.T) {
			_, _, err := Sign(Ed25519, []byte("private"), document)
			assert.EqualError(t, err, "ED25519 private key length must be 64, got 7")
		})
	})
}
