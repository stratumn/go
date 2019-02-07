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

package monitoring

import (
	"context"
	"encoding/hex"

	log "github.com/sirupsen/logrus"

	"go.elastic.co/apm"
)

const (
	txID = "apm_tx_id"
)

// LogEntry creates a new log entry with some default fields set.
func LogEntry() *log.Entry {
	return log.WithField("version", version).WithField("commit", commit)
}

// TxLogEntry creates a new log entry with the transaction ID set.
// This makes it easier to correlate transactions/spans with logs.
func TxLogEntry(ctx context.Context) *log.Entry {
	entry := LogEntry()
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		return entry
	}

	traceID := tx.TraceContext().Trace
	return entry.WithField(txID, hex.EncodeToString(traceID[:]))
}

// SpanLogEntry creates a new log entry with the transaction ID set.
// This makes it easier to correlate transactions/spans with logs.
func SpanLogEntry(span *apm.Span) *log.Entry {
	entry := LogEntry()
	if span.Dropped() {
		return entry
	}

	traceID := span.TraceContext().Trace
	return log.WithField(txID, hex.EncodeToString(traceID[:]))
}
