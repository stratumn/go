// Copyright 2017-2018 Stratumn SAS. All rights reserved.
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

package types_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stratumn/go-core/types"
)

func TestTransactionIDString(t *testing.T) {
	str := "8353334c6e4911e6ad927bd17dea491a"
	buf, _ := hex.DecodeString(str)
	txid := types.TransactionID(buf)

	if got, want := txid.String(), str; got != want {
		t.Errorf("txid.String() = %q want %q", got, want)
	}
}

func TestTransactionMarshalJSON(t *testing.T) {
	str := "8353334c6e4911e6ad927bd17dea491a"
	buf, _ := hex.DecodeString(str)
	txid := types.TransactionID(buf)
	marshalled, err := json.Marshal(txid)
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}

	if got, want := string(marshalled), fmt.Sprintf(`"%s"`, str); got != want {
		t.Errorf("txid.MarshalJSON() = %q want %q", got, want)
	}
}

func TestTransactionUnmarshalJSON(t *testing.T) {
	str := "8353334c6e4911e6ad927bd17dea491a"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var txid types.TransactionID
	err := json.Unmarshal([]byte(marshalled), &txid)
	if err != nil {
		t.Fatalf("json.Unmarshal(): err: %s", err)
	}

	if got, want := txid.String(), str; got != want {
		t.Errorf("txid.UnmarshalJSON() = %q want %q", got, want)
	}
}

func TestTransactionUnmarshalJSON_invalid(t *testing.T) {
	str := "azertyu"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var txid types.TransactionID
	err := json.Unmarshal([]byte(marshalled), &txid)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}
