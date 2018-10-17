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

package gcloud

import (
	"encoding/base64"
	"errors"
	"os"

	"google.golang.org/api/option"
)

// Errors used by the configuration module.
var (
	ErrInvalidCredentials = errors.New("GOOGLE_APPLICATION_CREDENTIALS_DATA must contain the content of a service account file as a base64 string")
)

// GetCredentials looks through the environment for a way to
// authenticate to a service using 'application default credentials'.
func GetCredentials() (option.ClientOption, error) {
	if credentialsFile, found := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); found {
		return option.WithCredentialsFile(credentialsFile), nil
	}
	if credentialsB64, found := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS_DATA"); found {
		credentialsJSON, err := base64.StdEncoding.DecodeString(credentialsB64)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
		return option.WithCredentialsJSON(credentialsJSON), nil
	}
	return nil, nil
}
