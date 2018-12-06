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

package blockcypher

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	ctx := context.Background()
	waitRetryBase = time.Millisecond

	t.Run("success", func(t *testing.T) {
		err := RetryWithBackOff(ctx, func() error { return nil })
		require.NoError(t, err)
	})

	t.Run("success after retry", func(t *testing.T) {
		i := 0

		err := RetryWithBackOff(ctx, func() error {
			if i == 3 {
				return nil
			}

			i++
			return errors.New("failed")
		})

		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		err := RetryWithBackOff(ctx, func() error { return errors.New("fatal") })
		require.EqualError(t, err, "fatal")
	})
}
