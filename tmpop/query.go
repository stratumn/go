// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import "encoding/json"

// Query types.
const (
	GetInfo      = "GetInfo"
	GetSegment   = "GetSegment"
	FindSegments = "FindSegments"
	GetMapIDs    = "GetMapIDs"
	GetValue     = "GetValue"
)

// BuildQueryBinary outputs the marshalled Query.
func BuildQueryBinary(args interface{}) (argsBytes []byte, err error) {
	if args != nil {
		if argsBytes, err = json.Marshal(args); err != nil {
			return
		}
	}
	return
}
