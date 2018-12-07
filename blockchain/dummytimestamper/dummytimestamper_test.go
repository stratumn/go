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

package dummytimestamper_test

import (
	"context"
	"testing"

	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/blockchain/dummytimestamper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummyTimestamper(t *testing.T) {
	ctx := context.Background()
	ts := &dummytimestamper.Timestamper{}

	t.Run("network string", func(t *testing.T) {
		n := dummytimestamper.Network{}
		assert.Equal(t, "dummytimestamper", n.String())
	})

	t.Run("timestamper network string", func(t *testing.T) {
		assert.Equal(t, "dummytimestamper", ts.Network().String())
	})

	t.Run("timestamp", func(t *testing.T) {
		_, err := ts.Timestamp(ctx, map[string][]byte{"hash": chainscripttest.RandomHash()})
		require.NoError(t, err)
	})

	t.Run("timestamp hash", func(t *testing.T) {
		_, err := ts.TimestampHash(ctx, chainscripttest.RandomHash())
		require.NoError(t, err)
	})
}
