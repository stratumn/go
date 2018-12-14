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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewSession creates a new session in the given AWS region.
// If you run locally, you need valid credentials in the ~/.aws folder or in
// your environment variables.
// If you run on an AWS EC2 instance of inside ECS, you just need to configure
// an IAM role for your instance that has the necessary permissions and
// credentials will be fetched automatically.
func NewSession(region string) (*session.Session, error) {
	return session.NewSession(
		aws.NewConfig().WithRegion(region),
	)
}
