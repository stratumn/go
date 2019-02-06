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

package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stratumn/go-core/cloud/aws"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/fossilizer/dummyqueue"
	"github.com/stratumn/go-core/monitoring"
)

// Queues that can be used by the BTC fossilizer.
const (
	DummyQueue = "dummy"
	AWSQueue   = "aws"
)

// Default values for command-line flags.
const (
	DefaultFossilsQueue = "pending-fossils"
)

// QueueFromFlags creates a fossils queue from command-line flags.
func QueueFromFlags() fossilizer.FossilsQueue {
	switch *queueType {
	case DummyQueue:
		return dummyqueue.New()
	case AWSQueue:
		session := aws.SessionFromFlags()
		client := sqs.New(session)
		queueURL := QueueURL(client, fossilsQueue)
		return aws.NewFossilsQueue(client, queueURL)
	default:
		monitoring.LogEntry().WithField("queueType", *queueType).Fatal("unknown queue type")
		return nil
	}
}

// QueueURL retrieves the queue URL from its name.
func QueueURL(client *sqs.SQS, queueName *string) *string {
	r, err := client.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: queueName})
	if err != nil {
		monitoring.LogEntry().WithField("error", err.Error()).Fatal("aws configuration is invalid")
	}

	return r.QueueUrl
}
