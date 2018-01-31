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

package validator

import (
	"encoding/base64"
	"strings"

	cj "github.com/gibson042/canonicaljson-go"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"golang.org/x/crypto/ed25519"
)

const (
	// Ed25519 is the EdDSA signature scheme using SHA-512/256 and Curve25519
	Ed25519 = "ed25519"
)

var (
	// SupportedSignatureTypes is a list of the supported signature types
	SupportedSignatureTypes = []string{Ed25519}

	// ErrMissingSignature is returned when there are no signatures in the link
	ErrMissingSignature = errors.New("signature validation requires link.signatures to contain at least one element")

	// ErrInvalidSignature is returned when the siganture verifcation failed
	ErrInvalidSignature = errors.New("signature verification failed")

	// ErrUnsupportedSignatureType is returned when the signature type is not supported
	ErrUnsupportedSignatureType = errors.Errorf("signature type must be one of %v", SupportedSignatureTypes)

	// ErrEmptyPayload is returned when the JMESPATH query didn't match any element of the link
	ErrEmptyPayload = errors.New("JMESPATH query does not match any link data")
)

// signatureValidatorConfig contains everything a signatureValidator needs to
// validate links.
type signatureValidatorConfig struct {
	*validatorBaseConfig
}

// newSignatureValidatorConfig creates a signatureValidatorConfig for a given process and type.
func newSignatureValidatorConfig(process, linkType string) (*signatureValidatorConfig, error) {
	baseConfig, err := newValidatorBaseConfig(process, linkType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &signatureValidatorConfig{baseConfig}, nil
}

// signatureValidator validates the json signature of a link's state.
type signatureValidator struct {
	config *signatureValidatorConfig
}

func newSignatureValidator(config *signatureValidatorConfig) validator {
	return &signatureValidator{config: config}
}

func (sv signatureValidator) isSignatureSupported(sigType string) bool {
	for _, supportedSig := range SupportedSignatureTypes {
		if strings.ToLower(sigType) == supportedSig {
			return true
		}
	}
	return false
}

// Validate validates the signature of a link's state.
func (sv signatureValidator) Validate(_ store.SegmentReader, link *cs.Link) error {
	if !sv.config.shouldValidate(link) {
		return nil
	}

	if len(link.Signatures) == 0 {
		return ErrMissingSignature
	}

	for _, sig := range link.Signatures {
		if !sv.isSignatureSupported(sig.Type) {
			return ErrUnsupportedSignatureType
		}

		// don't check decoding errors here, this is done in link.Validate() beforehand
		bytesKey, _ := base64.StdEncoding.DecodeString(sig.PublicKey)
		bytesSig, _ := base64.StdEncoding.DecodeString(sig.Signature)

		payload, err := jmespath.Search(sig.Payload, link)
		if err != nil {
			return errors.Errorf("failed to execute jmespath query : %s", err.Error())
		}
		if payload == nil {
			return ErrEmptyPayload
		}

		payloadBytes, err := cj.Marshal(payload)
		if err != nil {
			return err
		}

		switch sig.Type {
		case Ed25519:
			publicKey := ed25519.PublicKey(bytesKey)
			if len(publicKey) != ed25519.PublicKeySize {
				return errors.Errorf("Ed25519 public key lenght must be %d, got %d", ed25519.PublicKeySize, len(publicKey))
			}
			if !ed25519.Verify(publicKey, payloadBytes, bytesSig) {
				return ErrInvalidSignature
			}
		}
	}
	// TODO: check that
	// - public keys match PKI of rules.json
	// - required signatures for this action are present/valid

	return nil
}
