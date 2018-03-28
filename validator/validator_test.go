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

package validator_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stratumn/go-crypto/keys"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/utils"
	"github.com/stratumn/go-indigocore/validator"
)

type testCase struct {
	name  string
	link  *cs.Link
	valid bool
}

const (
	AlicePrivateKey = "-----BEGIN ED25519 PRIVATE KEY-----\nBEC0TyVE2Y7+OgPHcSAAIAjUHCVA68swAp235LkQZBIrZnUfW/lss95djRXjIeX+\nezH5bdbVe7s4wbPJRBiej+it\n-----END ED25519 PRIVATE KEY-----\n"
)

func initTestCases(t *testing.T) (store.Adapter, []testCase) {
	store := dummystore.New(nil)
	initAuctionLink := &cs.Link{
		State: map[string]interface{}{
			"buyer":        "alice",
			"seller":       "bob",
			"lot":          "painting",
			"initialPrice": 12,
		},
		Meta: cs.LinkMeta{
			Process: "auction",
			Type:    "init",
		},
	}
	priv, _, err := keys.ParseSecretKey([]byte(AlicePrivateKey))
	initAuctionLinkHash, err := store.CreateLink(context.Background(), cstesting.SignLinkWithKey(initAuctionLink, priv))
	require.NoError(t, err)

	var testCases = []testCase{{
		name:  "valid-init-link",
		link:  initAuctionLink,
		valid: true,
	}, {
		name: "valid-link",
		link: &cs.Link{
			State: map[string]interface{}{
				"buyer":    "alice",
				"bidPrice": 42,
			},
			Meta: cs.LinkMeta{
				Process:      "auction",
				PrevLinkHash: initAuctionLinkHash.String(),
				Type:         "bid",
			},
		},
		valid: true,
	}, {
		name: "no-validator-match",
		link: &cs.Link{
			Meta: cs.LinkMeta{
				Process: "auction",
				Type:    "unknown",
			},
		},
		valid: false,
	}, {
		name: "missing-required-field",
		link: &cs.Link{
			State: map[string]interface{}{
				"to": "bob",
			},
			Meta: cs.LinkMeta{
				Process: "chat",
				Type:    "message",
			},
		},
		valid: false,
	}, {
		name: "invalid-field-value",
		link: &cs.Link{
			State: map[string]interface{}{
				"buyer":    "alice",
				"bidPrice": -10,
			},
			Meta: cs.LinkMeta{
				Process: "auction",
				Type:    "bid",
			},
		},
		valid: false,
	}}
	return store, testCases
}

func TestValidator(t *testing.T) {
	testFile := utils.CreateTempFile(t, validator.ValidJSONConfig)
	defer os.Remove(testFile)

	children, err := validator.LoadConfig(testFile, nil)
	assert.NoError(t, err, "LoadConfig()")

	v := validator.NewMultiValidator(children)

	store, testCases := initTestCases(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(context.Background(), store, tt.link)
			if tt.valid {
				assert.NoError(t, err, "v.Validate()")
			} else {
				assert.Error(t, err, "v.Validate()")
			}
		})
	}
}
