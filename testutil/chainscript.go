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

package testutil

import (
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
)

// RandomLink creates a link with random data.
func RandomLink(t *testing.T) *chainscript.Link {
	return chainscripttest.NewLinkBuilder(t).WithRandomData().Build()
}

// RandomSegment creates a segment with random data.
func RandomSegment(t *testing.T) *chainscript.Segment {
	return chainscripttest.NewLinkBuilder(t).WithRandomData().Segmentify(t)
}

// RandomEvidence creates a random evidence.
func RandomEvidence(t *testing.T) *chainscript.Evidence {
	return &chainscript.Evidence{
		Version:  "1.0.0",
		Backend:  chainscripttest.RandomString(6),
		Provider: chainscripttest.RandomString(10),
		Proof:    chainscripttest.RandomBytes(64),
	}
}
