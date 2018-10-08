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

package types

import (
	"encoding/hex"
	"encoding/json"
)

// TransactionID is a blockchain transaction ID.
type TransactionID []byte

// String returns a hex encoded string.
func (txid TransactionID) String() string {
	return hex.EncodeToString(txid)
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (txid TransactionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(txid.String())
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (txid *TransactionID) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	*txid, err = hex.DecodeString(s)
	return
}
