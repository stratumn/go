// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

func TestTransactionIDString(t *testing.T) {
	str := "8353334c6e4911e6ad927bd17dea491a"
	buf, _ := hex.DecodeString(str)
	txid := TransactionID(buf)

	if got, want := txid.String(), str; got != want {
		t.Errorf("tix.String() = %q want %q", got, want)
	}
}

func TestTransactionMarshalJSON(t *testing.T) {
	str := "8353334c6e4911e6ad927bd17dea491a"
	buf, _ := hex.DecodeString(str)
	txid := TransactionID(buf)
	marshalled, err := json.Marshal(txid)
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}

	if got, want := string(marshalled), fmt.Sprintf(`"%s"`, str); got != want {
		t.Errorf("tix.MarshalJSON() = %q want %q", got, want)
	}
}

func TestTransactionUnmarshalJSON(t *testing.T) {
	str := "8353334c6e4911e6ad927bd17dea491a"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var txid TransactionID
	err := json.Unmarshal([]byte(marshalled), &txid)
	if err != nil {
		t.Fatalf("json.Unmarshal(): err: %s", err)
	}

	if got, want := txid.String(), str; got != want {
		t.Errorf("tix.UnmarshalJSON() = %q want %q", got, want)
	}
}

func TestTransactionUnmarshalJSON_invalid(t *testing.T) {
	str := "azertyu"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var txid TransactionID
	err := json.Unmarshal([]byte(marshalled), &txid)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}
