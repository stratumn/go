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
	"github.com/stratumn/go-core/fossilizer/dummyexporter"
	"github.com/stratumn/go-core/monitoring"
)

// Exporters that can be used by the BTC fossilizer.
const (
	NoExporter      = ""
	ConsoleExporter = "console"
	AWSExporter     = "aws"
)

// Default values for command-line flags.
const (
	DefaultExporterQueue = "fossilizer-events"
)

// ExporterFromFlags creates a fossilizer events exporter from command-line
// flags.
func ExporterFromFlags() fossilizer.EventExporter {
	switch *exporter {
	case NoExporter:
		return nil
	case ConsoleExporter:
		return dummyexporter.New()
	case AWSExporter:
		session := aws.SessionFromFlags()
		client := sqs.New(session)
		queueURL := QueueURL(client, exporterQueue)
		return aws.NewEventExporter(client, queueURL)
	default:
		monitoring.LogEntry().WithField("exporter", *exporter).Fatal("unknown exporter type")
		return nil
	}
}
