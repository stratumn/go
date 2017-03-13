// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import "encoding/json"

// Query types.
const (
	GetInfo       = "GetInfo"
	GetSegment    = "GetSegment"
	FindSegments  = "FindSegments"
	GetMapIDs     = "GetMapIDs"
	DeleteSegment = "DeleteSegment"
)

// Query is the type used to query the tendermint App
type Query struct {
	Name string `json:"Name"`
	Args []byte `json:"Args"`
}

// BuildQueryBinary outputs the marshalled Query
func BuildQueryBinary(name string, args interface{}) ([]byte, error) {
	var argsBytes []byte
	if args != nil {
		var err error
		if argsBytes, err = json.Marshal(args); err != nil {
			return nil, err
		}
	}

	query := &Query{Name: name, Args: argsBytes}
	bytes, err := json.Marshal(query)

	if err != nil {
		return nil, err
	}
	return bytes, nil
}
