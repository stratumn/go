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
	"context"
	"testing"

	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/testutil"

	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	process = "test"

	stateCreatedProduct     = "createdProduct"
	stateSignedProduct      = "signedProduct"
	stateFinalProduct       = "finalProduct"
	stateFinalSignedProduct = "finalSignedProduct"
	stateHackedProduct      = "hackedProduct"
)

var (
	// Allowed transitions
	createdProductTransitions     = []string{""}
	signedProductTransitions      = []string{stateCreatedProduct}
	finalProductTransitions       = []string{stateSignedProduct}
	hackedProductTransitions      = []string{stateCreatedProduct, stateHackedProduct}
	finalSignedProductTransitions = []string{stateFinalProduct, stateCreatedProduct, stateHackedProduct}
)

type stateMachineLinks struct {
	createdProduct *cs.Link

	// Complete workflow
	signedProduct      *cs.Link
	finalProduct       *cs.Link
	finalSignedProduct *cs.Link

	// Hacked workflow
	hacked1Product     *cs.Link
	hacked2Product     *cs.Link
	hackedFinalProduct *cs.Link

	// Bypass workflow
	approvedProduct *cs.Link
}

func populateStore(t *testing.T) (store.Adapter, stateMachineLinks) {
	store := dummystore.New(nil)
	require.NotNil(t, store)

	var links stateMachineLinks

	links.createdProduct = cstesting.NewLinkBuilder().
		WithProcess(process).
		WithType(stateCreatedProduct).
		WithPrevLinkHash("").
		Build()
	_, err := store.CreateLink(context.Background(), links.createdProduct)
	require.NoError(t, err)

	appendLink := func(prevLink *cs.Link, linkType string) *cs.Link {
		l := cstesting.NewLinkBuilder().
			Branch(prevLink).
			WithType(linkType).
			Build()
		_, err := store.CreateLink(context.Background(), l)
		require.NoError(t, err)
		return l
	}

	links.signedProduct = appendLink(links.createdProduct, stateSignedProduct)
	links.finalProduct = appendLink(links.signedProduct, stateFinalProduct)
	links.finalSignedProduct = appendLink(links.finalProduct, stateFinalSignedProduct)

	links.approvedProduct = appendLink(links.createdProduct, stateFinalSignedProduct)

	links.hacked1Product = appendLink(links.createdProduct, stateHackedProduct)
	links.hacked2Product = appendLink(links.hacked1Product, stateHackedProduct)
	links.hackedFinalProduct = appendLink(links.hacked2Product, stateFinalSignedProduct)
	return store, links
}

func TestTransitionValidator(t *testing.T) {
	t.Parallel()
	store, links := populateStore(t)

	type testCase struct {
		name        string
		valid       bool
		err         string
		link        *cs.Link
		transitions allowedTransitions
	}

	testCases := []testCase{
		{
			name:        "good init",
			valid:       true,
			link:        links.createdProduct,
			transitions: createdProductTransitions,
		},
		{
			name:        "bad init",
			valid:       false,
			err:         `no transition found \(\) --> createdProduct`,
			link:        links.createdProduct,
			transitions: []string{"src1", "src2"},
		},
		{
			name:        "good final transition",
			valid:       true,
			link:        links.finalSignedProduct,
			transitions: finalSignedProductTransitions,
		},
		{
			name:        "good fast final transition",
			valid:       true,
			link:        links.approvedProduct,
			transitions: finalSignedProductTransitions,
		},
		{
			name:        "good hacked final transition",
			valid:       true,
			link:        links.hackedFinalProduct,
			transitions: finalSignedProductTransitions,
		},
		{
			name:        "invalid transition",
			valid:       false,
			err:         "no transition found finalProduct --> finalSignedProduct",
			link:        links.finalSignedProduct,
			transitions: []string{stateCreatedProduct},
		},
		{
			name:  "prevLink not found",
			valid: false,
			err:   "previous segment not found.*",
			link: func() *cs.Link {
				l := cstesting.Clone(links.finalProduct)
				l.Meta.PrevLinkHash = testutil.RandomHash().String()
				return l
			}(),
			transitions: []string{stateCreatedProduct},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			baseCfg, err := newValidatorBaseConfig(process, tt.link.Meta.Type)
			require.NoError(t, err)
			v := newTransitionValidator(baseCfg, tt.transitions)

			err = v.Validate(context.Background(), store, tt.link)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Regexp(t, tt.err, err.Error())
			}
		})
	}

}

func TestTransitionShouldValidate(t *testing.T) {
	t.Parallel()
	_, links := populateStore(t)

	type testCase struct {
		name string
		ret  bool
		conf *validatorBaseConfig
		link *cs.Link
	}

	newConf := func(process, linkType string) *validatorBaseConfig {
		cfg, err := newValidatorBaseConfig(process, linkType)
		require.NoError(t, err)
		return cfg
	}

	testCases := []testCase{
		{
			name: "has to validate",
			ret:  true,
			conf: newConf(links.createdProduct.Meta.Process, links.createdProduct.Meta.Type),
			link: links.createdProduct,
		},
		{
			name: "bad process",
			ret:  false,
			conf: newConf("foo", links.createdProduct.Meta.Type),
			link: links.createdProduct,
		},
		{
			name: "bad state",
			ret:  false,
			conf: newConf(links.createdProduct.Meta.Process, "bar"),
			link: links.createdProduct,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			v := newTransitionValidator(tt.conf, nil)
			assert.Equal(t, tt.ret, v.ShouldValidate(tt.link))
		})
	}

}

func TestTransitionHash(t *testing.T) {
	t.Parallel()
	_, links := populateStore(t)

	baseCfg, err := newValidatorBaseConfig(process, links.finalProduct.Meta.Type)
	require.NoError(t, err)
	v1 := newTransitionValidator(baseCfg, finalProductTransitions)
	v2 := newTransitionValidator(baseCfg, createdProductTransitions)

	hash1, err1 := v1.Hash()
	hash2, err2 := v2.Hash()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, hash1)
	assert.NotNil(t, hash2)
	assert.NotEqual(t, hash1.String(), hash2.String())
}
