// Copyright 2017 Stratumn SAS. All rights reserved.
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

// Package types defines common types.
package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Bytes20Size is the size of a 20-byte long byte array.
const Bytes20Size = 20

// Bytes20 is a 20-byte long byte array.
type Bytes20 [Bytes20Size]byte

// NewBytes20FromString creates a Bytes20 from a hex encoded string.
func NewBytes20FromString(src string) (*Bytes20, error) {
	var b Bytes20
	if err := b.Unstring(src); err != nil {
		return nil, err
	}
	return &b, nil
}

// String returns a hex encoded string.
func (b *Bytes20) String() string {
	return hex.EncodeToString(b[:])
}

// Unstring sets the value from a hex encoded string.
func (b *Bytes20) Unstring(src string) error {
	buf, err := hex.DecodeString(src)
	if err != nil {
		return err
	}
	if n := len(buf); n != Bytes20Size {
		return fmt.Errorf("invalid Bytes20 size got %d want %d", n, Bytes20Size)
	}

	copy(b[:], buf)
	return nil
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (b *Bytes20) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (b *Bytes20) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return b.Unstring(s)
}

// Reverse reverses the bytes order.
func (b *Bytes20) Reverse(rb *ReversedBytes20) {
	i := Bytes20Size - 1
	for _, v := range b {
		rb[i] = v
		i--
	}
}

// ReversedBytes20 is a 20-byte long byte reversed array.
// While the bytes are reversed, the hex encoded strings are not.
type ReversedBytes20 [Bytes20Size]byte

// NewReversedBytes20FromString creates a ReversedBytes20 from a hex encoded
// string.
func NewReversedBytes20FromString(src string) (*ReversedBytes20, error) {
	b, err := NewBytes20FromString(src)
	if err != nil {
		return nil, err
	}
	var rb ReversedBytes20
	b.Reverse(&rb)
	return &rb, nil
}

// String returns a hex encoded string.
func (rb *ReversedBytes20) String() string {
	var b Bytes20
	rb.Reverse(&b)
	return b.String()
}

// Unstring sets the value from a hex encoded string.
func (rb *ReversedBytes20) Unstring(src string) error {
	b, err := NewBytes20FromString(src)
	if err != nil {
		return err
	}
	b.Reverse(rb)
	return nil
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (rb *ReversedBytes20) MarshalJSON() ([]byte, error) {
	var b Bytes20
	rb.Reverse(&b)
	return b.MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (rb *ReversedBytes20) UnmarshalJSON(data []byte) error {
	var b Bytes20
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	b.Reverse(rb)
	return nil
}

// Reverse reverses the bytes order.
func (rb *ReversedBytes20) Reverse(b *Bytes20) {
	i := Bytes20Size - 1
	for _, v := range rb {
		b[i] = v
		i--
	}
}

// Bytes32Size is the size of a 32-byte long byte array.
const Bytes32Size = 32

// Bytes32Zero is the default value for a 32-byte long byte array.
var Bytes32Zero = &Bytes32{}

// Bytes32 is a 32-byte long byte array.
type Bytes32 [Bytes32Size]byte

// NewBytes32FromString creates a Bytes32 from a hex encoded string.
func NewBytes32FromString(src string) (*Bytes32, error) {
	var b Bytes32
	if err := b.Unstring(src); err != nil {
		return nil, err
	}
	return &b, nil
}

// NewBytes32FromBytes creates a Bytes32 from a byte slice.
func NewBytes32FromBytes(src []byte) *Bytes32 {
	var b Bytes32
	if len(src) > 0 {
		copy(b[:], src)
	}

	return &b
}

// String returns a hex encoded string.
func (b *Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

// Unstring sets the value from a hex encoded string.
func (b *Bytes32) Unstring(src string) error {
	buf, err := hex.DecodeString(src)
	if err != nil {
		return err
	}
	if n := len(buf); n != Bytes32Size {
		return fmt.Errorf("invalid Bytes32 size got %d want %d", n, Bytes32Size)
	}

	copy(b[:], buf)
	return nil
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (b *Bytes32) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (b *Bytes32) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return b.Unstring(s)
}

// Compare compares two Bytes32
func (b *Bytes32) Compare(b2 *Bytes32) int {
	if b.Zero() || b2.Zero() {
		if b.Zero() && b2.Zero() {
			return 0
		}

		return 1
	}

	return bytes.Compare(b[:], b2[:])
}

// Equals checks if two Bytes32 are equal
func (b *Bytes32) Equals(b2 *Bytes32) bool {
	return b.Compare(b2) == 0
}

// EqualsBytes checks if a byte slice equals a Bytes32
func (b *Bytes32) EqualsBytes(b2 []byte) bool {
	if len(b2) == 0 && b.Zero() {
		return true
	}

	if b == nil {
		return false
	}

	return bytes.Equal(b[:], b2)
}

// Zero checks if a Bytes32 is the default value or nil
func (b *Bytes32) Zero() bool {
	return b == nil || bytes.Equal(b[:], Bytes32Zero[:])
}

// Reverse reverses the bytes order.
func (b *Bytes32) Reverse(rb *ReversedBytes32) {
	for i, v := range b {
		rb[Bytes32Size-i-1] = v
	}
}

// Value implements the database.sql.driver.Valuer interface.
func (b *Bytes32) Value() (driver.Value, error) {
	return fmt.Sprintf("\\x%s", b.String()), nil
}

// ReversedBytes32 is a 32-byte long byte reversed array.
// While the bytes are reversed, the hex encoded strings are not.
type ReversedBytes32 [Bytes32Size]byte

// NewReversedBytes32FromString creates a ReversedBytes32 from a hex encoded
// string.
func NewReversedBytes32FromString(src string) (*ReversedBytes32, error) {
	b, err := NewBytes32FromString(src)
	if err != nil {
		return nil, err
	}
	var rb ReversedBytes32
	b.Reverse(&rb)
	return &rb, nil
}

// String returns a hex encoded string.
func (rb *ReversedBytes32) String() string {
	var b Bytes32
	rb.Reverse(&b)
	return b.String()
}

// Unstring sets the value from a hex encoded string.
func (rb *ReversedBytes32) Unstring(src string) error {
	b, err := NewBytes32FromString(src)
	if err != nil {
		return err
	}
	b.Reverse(rb)
	return nil
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (rb *ReversedBytes32) MarshalJSON() ([]byte, error) {
	var b Bytes32
	rb.Reverse(&b)
	return b.MarshalJSON()
}

// UnmarshalJSON implements encoding/json.Unmarshaler.UnmarshalJSON.
func (rb *ReversedBytes32) UnmarshalJSON(data []byte) error {
	var b Bytes32
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	b.Reverse(rb)
	return nil
}

// Reverse reverses the bytes order.
func (rb *ReversedBytes32) Reverse(b *Bytes32) {
	for i, v := range rb {
		b[Bytes32Size-i-1] = v
	}
}

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
