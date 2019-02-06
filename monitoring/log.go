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
	noTx = "not sampled"
)

// LogWithTxFields adds transaction fields to a log entry.
// This makes it easier to correlate transactions/spans with logs.
func LogWithTxFields(ctx context.Context) *log.Entry {
	tx := apm.TransactionFromContext(ctx)
	if tx != nil {
		traceID := tx.TraceContext().Trace
		return log.WithField(txID, hex.EncodeToString(traceID[:]))
	}

	return log.WithField(txID, noTx)
}

// LogWithSpanFields adds span fields to a log entry.
// This makes it easier to correlate transactions/spans with logs.
func LogWithSpanFields(span *apm.Span) *log.Entry {
	if !span.Dropped() {
		traceID := span.TraceContext().Trace
		return log.WithField(txID, hex.EncodeToString(traceID[:]))
	}

	return log.WithField(txID, noTx)
}
