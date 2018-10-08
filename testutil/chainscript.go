// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/types"
	"github.com/stretchr/testify/require"
)

// PaginatedSegmentsEqual verifies that two paginated segment lists are equal.
func PaginatedSegmentsEqual(t *testing.T, s1, s2 *types.PaginatedSegments) {
	require.Equal(t, s1.TotalCount, s2.TotalCount, "TotalCount")
	require.Equal(t, len(s1.Segments), len(s2.Segments), "Length")

	for i, s := range s1.Segments {
		chainscripttest.SegmentsEqual(t, s, s2.Segments[i])
	}
}
