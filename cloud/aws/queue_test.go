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

package aws_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	awsqueue "github.com/stratumn/go-core/cloud/aws"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testQueuePrefix = "go-core-ci-test-"
	testRegion      = "eu-west-3"
)

func TestFossilsQueue(t *testing.T) {
	sess, err := awsqueue.NewSession(testRegion)
	require.NoError(t, err)

	client := sqs.New(sess)
	defer cleanTestQueues(t, client)

	t.Run("push and pop", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		queueURL := createTestQueue(t, client)
		q := awsqueue.NewFossilsQueue(client, queueURL)

		fossilsCount := byte(5)
		for i := byte(0); i < fossilsCount; i++ {
			err := q.Push(ctx, &fossilizer.Fossil{
				Data: []byte{i},
				Meta: []byte{i + 10},
			})
			require.NoError(t, err)
		}

		var fossils []*fossilizer.Fossil
		for i := byte(0); i < fossilsCount; i++ {
			f, err := q.Pop(ctx, 1)
			require.NoError(t, err)
			require.Len(t, f, 1)
			fossils = append(fossils, f...)
		}

		verifyFossils(t, 5, fossils)
	})

	t.Run("pop empty queue", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		queueURL := createTestQueue(t, client)
		q := awsqueue.NewFossilsQueue(client, queueURL)

		fossils, err := q.Pop(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, fossils, 0)
	})

	t.Run("pop many", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		queueURL := createTestQueue(t, client)
		q := awsqueue.NewFossilsQueue(client, queueURL)

		for i := byte(0); i < 15; i++ {
			err := q.Push(ctx, &fossilizer.Fossil{
				Data: []byte{i},
				Meta: []byte{i + 50},
			})
			require.NoError(t, err)
		}

		fossils1, err := q.Pop(ctx, 13)
		require.NoError(t, err)
		require.Len(t, fossils1, 13)

		fossils2, err := q.Pop(ctx, 2)
		require.NoError(t, err)
		require.Len(t, fossils2, 2)

		verifyFossils(t, 15, append(fossils1, fossils2...))
	})

	t.Run("pop more than queue size", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		queueURL := createTestQueue(t, client)
		q := awsqueue.NewFossilsQueue(client, queueURL)

		for i := byte(0); i < 25; i++ {
			err := q.Push(ctx, &fossilizer.Fossil{
				Data: []byte{i},
				Meta: []byte{i + 50},
			})
			require.NoError(t, err)
		}

		fossils, err := q.Pop(ctx, 32*1024)
		require.NoError(t, err)
		require.Len(t, fossils, 25)

		verifyFossils(t, 25, fossils)
	})
}

func verifyFossils(t *testing.T, expectedCount byte, fossils []*fossilizer.Fossil) {
	for i := byte(0); i < expectedCount; i++ {
		found := false
		for _, f := range fossils {
			if f.Data[0] == i {
				found = true
				break
			}
		}

		assert.Truef(t, found, "fossils %d not found", i)
	}
}

// createTestQueue creates a test queue with a random name.
// You need to manually delete the queue afterwards (see cleanTestQueues).
func createTestQueue(t *testing.T, client *sqs.SQS) *string {
	queueId, _ := rand.Int(rand.Reader, big.NewInt(1<<16))
	queueName := fmt.Sprintf("%s%d", testQueuePrefix, queueId.Uint64())

	r, err := client.CreateQueue(&sqs.CreateQueueInput{QueueName: &queueName})
	require.NoError(t, err)

	return r.QueueUrl
}

// cleanTestQueues deletes all test queues from AWS.
// We don't want resources to be created and forgotten because we pay for it.
func cleanTestQueues(t *testing.T, client *sqs.SQS) {
	prefix := testQueuePrefix
	r, err := client.ListQueues(&sqs.ListQueuesInput{QueueNamePrefix: &prefix})
	assert.NoErrorf(t, err, "could not list test queues: you might need to manually delete test queues")

	for _, queueURL := range r.QueueUrls {
		_, err = client.DeleteQueue(&sqs.DeleteQueueInput{QueueUrl: queueURL})
		if err != nil {
			assert.Truef(t,
				strings.HasPrefix(err.Error(), sqs.ErrCodeQueueDoesNotExist),
				"could not delete test queue (%s): %s",
				*queueURL,
				err.Error,
			)
		}
	}
}
