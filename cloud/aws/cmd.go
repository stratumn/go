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

package aws

import (
	"flag"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stratumn/go-core/monitoring"
)

// Flags variables.
var (
	region string
)

// RegisterFlags registers the flags used by AWS components.
func RegisterFlags() {
	flag.StringVar(&region, "awsregion", "", "AWS region")
}

// SessionFromFlags returns an AWS session from command-line flags.
func SessionFromFlags() *session.Session {
	sess, err := NewSession(region)
	if err != nil {
		monitoring.LogEntry().WithField("error", err.Error()).Fatal("could not create AWS session")
	}

	return sess
}
