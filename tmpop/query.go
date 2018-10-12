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

package tmpop

import (
	"encoding/json"

	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/types"
)

// Query types.
const (
	AddEvidence   = "AddEvidence"
	FindSegments  = "FindSegments"
	GetEvidences  = "GetEvidences"
	GetInfo       = "GetInfo"
	GetMapIDs     = "GetMapIDs"
	GetSegment    = "GetSegment"
	PendingEvents = "PendingEvents"
)

// BuildQueryBinary outputs the marshalled Query.
func BuildQueryBinary(args interface{}) (argsBytes []byte, err error) {
	if args != nil {
		if argsBytes, err = json.Marshal(args); err != nil {
			return nil, types.WrapError(err, monitoring.InvalidArgument, Name, "json.Marshal")
		}
	}

	return
}
