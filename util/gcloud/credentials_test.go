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

package gcloud_test

import (
	"os"
	"testing"

	"github.com/stratumn/go-core/util/gcloud"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestGetCredentials(t *testing.T) {

	t.Run("with a file", func(t *testing.T) {
		os.Setenv(gcloud.CredentialsFileEnv, "test")
		defer os.Unsetenv(gcloud.CredentialsFileEnv)

		opt, err := gcloud.GetCredentials()
		require.Nil(t, err)
		require.Equal(t, option.WithCredentialsFile("test"), opt)

	})

	t.Run("with base64 credentials", func(t *testing.T) {
		os.Setenv(gcloud.CredentialsDataEnv, "dGVzdA==")
		defer os.Unsetenv(gcloud.CredentialsDataEnv)

		opt, err := gcloud.GetCredentials()
		require.Nil(t, err)
		require.Equal(t, option.WithCredentialsJSON([]byte("test")), opt)
	})

	t.Run("with invalid credentials", func(t *testing.T) {
		os.Setenv(gcloud.CredentialsDataEnv, "test?")
		defer os.Unsetenv(gcloud.CredentialsDataEnv)

		opt, err := gcloud.GetCredentials()
		require.EqualError(t, err, gcloud.ErrInvalidCredentials.Error())
		require.Nil(t, opt)
	})

	t.Run("without environment variables", func(t *testing.T) {
		opt, err := gcloud.GetCredentials()
		require.Nil(t, err)
		require.Nil(t, opt)
	})
}
