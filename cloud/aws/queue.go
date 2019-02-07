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
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"
)

const (
	// AWS documentation states that batches can be at most 10 messages.
	batchSize = 10

	// QueueComponent name for monitoring.
	QueueComponent = "FossilsQueue"
)

// FossilsQueue implements a fossils queue backed by AWS SQS.
type FossilsQueue struct {
	client   *sqs.SQS
	queueURL *string
}

// NewFossilsQueue connects to the given AWS SQS queue.
func NewFossilsQueue(client *sqs.SQS, queueURL *string) fossilizer.FossilsQueue {
	return &FossilsQueue{
		client:   client,
		queueURL: queueURL,
	}
}

// Push a fossil to the queue.
func (q *FossilsQueue) Push(ctx context.Context, f *fossilizer.Fossil) (err error) {
	span, _ := monitoring.StartSpanOutgoingRequest(ctx, "cloud/aws/queue/push")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	jsonFossil, err := json.Marshal(f)
	if err != nil {
		return types.WrapError(err, errorcode.InvalidArgument, QueueComponent, "json.Marshal")
	}

	body := string(jsonFossil)

	_, err = q.client.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    q.queueURL,
		MessageBody: &body,
	})
	if err != nil {
		return types.WrapError(err, errorcode.Unknown, QueueComponent, "error sending message to AWS SQS")
	}

	return nil
}

// Pop fossils from the queue.
func (q *FossilsQueue) Pop(ctx context.Context, count int) ([]*fossilizer.Fossil, error) {
	span, ctx := monitoring.StartSpanOutgoingRequest(ctx, "cloud/aws/queue/pop")
	defer span.End()

	var results []*fossilizer.Fossil
	for {
		nextBatchSize := batchSize
		if (count - len(results)) < nextBatchSize {
			nextBatchSize = count - len(results)
		}

		fossils, err := q.pop(ctx, int64(nextBatchSize))
		if err != nil {
			monitoring.SetSpanStatus(span, err)
			return nil, err
		}

		results = append(results, fossils...)

		// If the queue is empty, stop.
		if len(fossils) == 0 {
			return results, nil
		}

		// If we have the request count, stop.
		if len(results) == count {
			return results, nil
		}
	}
}

// pop a small number of elements (in the limits allowed by AWS).
func (q *FossilsQueue) pop(ctx context.Context, count int64) ([]*fossilizer.Fossil, error) {
	var fossils []*fossilizer.Fossil
	var entries []*sqs.DeleteMessageBatchRequestEntry

	r, err := q.client.ReceiveMessage(&sqs.ReceiveMessageInput{
		MaxNumberOfMessages: &count,
		QueueUrl:            q.queueURL,
	})
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unknown, QueueComponent, "error receiving message from AWS SQS")
	}

	// The queue is empty.
	if len(r.Messages) == 0 {
		return nil, nil
	}

	for _, m := range r.Messages {
		var fossil fossilizer.Fossil
		err := json.Unmarshal([]byte(*m.Body), &fossil)
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, QueueComponent, "json.Unmarshal")
		}

		fossils = append(fossils, &fossil)
		entries = append(entries, &sqs.DeleteMessageBatchRequestEntry{
			Id:            m.MessageId,
			ReceiptHandle: m.ReceiptHandle,
		})
	}

	_, err = q.client.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
		QueueUrl: q.queueURL,
		Entries:  entries,
	})
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unknown, QueueComponent, "error deleting message from AWS SQS")
	}

	return fossils, nil
}
