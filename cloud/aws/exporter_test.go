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
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	awsqueue "github.com/stratumn/go-core/cloud/aws"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventExporter(t *testing.T) {
	sess, err := awsqueue.NewSession(testRegion)
	require.NoError(t, err)

	client := sqs.New(sess)
	defer cleanTestQueues(t, client)

	t.Run("push to the queue", func(t *testing.T) {
		ctx := context.Background()
		queueURL := createTestQueue(t, client)
		exporter := awsqueue.NewEventExporter(client, queueURL)

		results := []*fossilizer.Result{
			newFossilizeResult([]byte{42}, []byte{24}),
			newFossilizeResult([]byte{43}, []byte{34}),
		}

		for _, r := range results {
			require.NoError(t, exporter.Push(ctx, &fossilizer.Event{
				EventType: fossilizer.DidFossilize,
				Data:      r,
			}))
		}

		msgResults := receiveFossilizeResults(t, client, queueURL, 2)
		assert.ElementsMatch(t, results, msgResults)
	})
}

func newFossilizeResult(data, meta []byte) *fossilizer.Result {
	return &fossilizer.Result{
		Fossil: fossilizer.Fossil{
			Data: data,
			Meta: meta,
		},
	}
}

// receiveFossilizeResults will poll the queue until the expected number of
// messages have been received.
func receiveFossilizeResults(t *testing.T, client *sqs.SQS, queueURL *string, count int64) []*fossilizer.Result {
	var results []*fossilizer.Result
	for {
		r, err := client.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            queueURL,
			MaxNumberOfMessages: &count,
		})
		require.NoError(t, err)

		if len(r.Messages) == 0 {
			return results
		}

		for _, m := range r.Messages {
			var msgEvent fossilizer.Event
			err := json.Unmarshal([]byte(*m.Body), &msgEvent)
			require.NoError(t, err)

			result, err := msgEvent.Result()
			require.NoError(t, err)

			results = append(results, result)
		}
	}
}
