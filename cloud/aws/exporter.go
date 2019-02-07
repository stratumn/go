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
	// ExporterComponent name for monitoring.
	ExporterComponent = "EventExporter"
)

// EventExporter exports fossilizer events to an AWS queue.
type EventExporter struct {
	client   *sqs.SQS
	queueURL *string
}

// NewEventExporter creates a new event exporter that sends events to a queue.
func NewEventExporter(client *sqs.SQS, queueURL *string) fossilizer.EventExporter {
	return &EventExporter{
		client:   client,
		queueURL: queueURL,
	}
}

// Push an event to the queue.
func (e *EventExporter) Push(ctx context.Context, event *fossilizer.Event) (err error) {
	span, _ := monitoring.StartSpanOutgoingRequest(ctx, "cloud/aws/exporter/push")
	defer func() {
		monitoring.SetSpanStatusAndEnd(span, err)
	}()

	jsonEvent, err := json.Marshal(event)
	if err != nil {
		return types.WrapError(err, errorcode.InvalidArgument, ExporterComponent, "json.Marshal")
	}

	body := string(jsonEvent)
	_, err = e.client.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    e.queueURL,
		MessageBody: &body,
	})
	if err != nil {
		return types.WrapError(err, errorcode.Unknown, ExporterComponent, "error sending message to AWS SQS")
	}

	return nil
}
