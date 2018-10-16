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
	"fmt"

	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/types"

	"go.opencensus.io/trace"
)

// SetSpanStatusAndEnd sets the status of the span depending on the error
// and ends it.
// You should usually call:
// defer func() {
//     SetSpanStatusAndEnd(span, err)
// }()
func SetSpanStatusAndEnd(span *trace.Span, err error) {
	SetSpanStatus(span, err)
	span.End()
}

// SetSpanStatus sets the status of the span depending on the error.
func SetSpanStatus(span *trace.Span, err error) {
	if err != nil {
		switch e := err.(type) {
		case *types.Error:
			span.AddAttributes(
				trace.Int64Attribute("error code", int64(e.Code)),
				trace.StringAttribute("component", e.Component),
			)
			span.SetStatus(trace.Status{
				Code: int32(e.Code),
				// We want to include a stack trace to make it easy to
				// investigate, hence the format.
				Message: fmt.Sprintf("%v+", e),
			})
		default:
			span.SetStatus(trace.Status{
				Code:    errorcode.Unknown,
				Message: err.Error(),
			})
		}
	}
}
