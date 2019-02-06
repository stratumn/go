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
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"

	"go.elastic.co/apm"
)

// Span types.
const (
	SpanTypeIncomingRequest = "app.request.incoming"
	SpanTypeOutgoingRequest = "app.request.outgoing"
	SpanTypeProcessing      = "app.processing"
)

// SetSpanStatusAndEnd sets the status of the span depending on the error
// and ends it.
// You should usually call:
// defer func() {
//     SetSpanStatusAndEnd(span, err)
// }()
func SetSpanStatusAndEnd(span *apm.Span, err error) {
	SetSpanStatus(span, err)
	span.End()
}

// SetSpanStatus sets the status of the span depending on the error.
func SetSpanStatus(span *apm.Span, err error) {
	if err != nil {
		switch e := err.(type) {
		case *types.Error:
			// We want to include a stack trace to make it easy to
			// investigate, hence the format.
			SpanLogEntry(span).Errorf("%v+", e)

			span.Context.SetTag(ErrorLabel, e.Error())
			span.Context.SetTag(ErrorCodeLabel, errorcode.Text(e.Code))
			span.Context.SetTag(ErrorComponentLabel, e.Component)
		default:
			SpanLogEntry(span).Errorf("%v+", err)
			span.Context.SetTag(ErrorLabel, err.Error())
		}
	}
}
