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

package types_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stratumn/sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBytes20FromString(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	b, err := types.NewBytes20FromString(str)
	if err != nil {
		t.Fatalf("NewBytes20FromString(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestNewBytes20FromString_invalidHex(t *testing.T) {
	if _, err := types.NewBytes20FromString("z234567890123456789012345678901234567890"); err == nil {
		t.Error("NewBytes20FromString(): err = nil want Error")
	}
}

func TestBytes20String(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	buf, _ := hex.DecodeString(str)
	var b types.Bytes20
	copy(b[:], buf)

	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestBytes20Unstring(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	var b types.Bytes20
	if err := b.Unstring(str); err != nil {
		t.Fatalf("b.Unstring(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestBytes20Unstring_invalidHex(t *testing.T) {
	var b types.Bytes20
	if err := b.Unstring("123456789012345678901234567890123456789q"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestBytes20Unstring_invalidSize(t *testing.T) {
	var b types.Bytes20
	if err := b.Unstring("12345678901234567890"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestBytes20MarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	buf, _ := hex.DecodeString(str)
	var b types.Bytes20
	copy(b[:], buf)
	marshalled, err := json.Marshal(&b)
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}

	if got, want := string(marshalled), fmt.Sprintf(`"%s"`, str); got != want {
		t.Errorf("b.MarshalJSON() = %q want %q", got, want)
	}
}

func TestBytes20UnmarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.Bytes20
	err := json.Unmarshal([]byte(marshalled), &b)
	if err != nil {
		t.Fatalf("json.Unmarshal(): err: %s", err)
	}

	if got, want := b.String(), str; got != want {
		t.Errorf("b.UnmarshalJSON() = %q want %q", got, want)
	}
}

func TestBytes20UnmarshalJSON_invalidStr(t *testing.T) {
	marshalled, err := json.Marshal([]string{"test"})
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}
	var b types.Bytes20
	err = json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestBytes20UnmarshalJSON_invalidHex(t *testing.T) {
	str := "+234567890123456789012345678901234567890"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.Bytes20
	err := json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestBytes20Reverse(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	buf, _ := hex.DecodeString(str)
	var b types.Bytes20
	copy(b[:], buf)
	var rev types.ReversedBytes20
	b.Reverse(&rev)

	for i := range rev {
		if got, want := rev[i], b[len(b)-i-1]; got != want {
			t.Errorf("rev[%d] = %x want %x", i, got, want)
		}
	}
}

func TestNewReversedBytes20FromString(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	b, err := types.NewReversedBytes20FromString(str)
	if err != nil {
		t.Fatalf("NewReversedBytes20FromString(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestNewReversedBytes20FromString_invalidHex(t *testing.T) {
	if _, err := types.NewReversedBytes20FromString("z234567890123456789012345678901234567890"); err == nil {
		t.Error("NewReversedBytes20FromString(): err = nil want Error")
	}
}

func TestReversedBytes20String(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	revStr := "9078563412907856341290785634129078563412"
	buf, _ := hex.DecodeString(str)
	var b types.ReversedBytes20
	copy(b[:], buf)

	if got, want := b.String(), revStr; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestReversedBytes20Unstring(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	var b types.ReversedBytes20
	if err := b.Unstring(str); err != nil {
		t.Fatalf("b.Unstring(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestReversedBytes20Unstring_invalidHex(t *testing.T) {
	var b types.ReversedBytes20
	if err := b.Unstring("u234567890123456789012345678901234567890"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestReversedBytes20Unstring_invalidSize(t *testing.T) {
	var b types.ReversedBytes20
	if err := b.Unstring("12345678901245678901234567890123456901234567891234567890234567"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestReversedBytes20MarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	revStr := "9078563412907856341290785634129078563412"
	buf, _ := hex.DecodeString(str)
	var b types.ReversedBytes20
	copy(b[:], buf)
	marshalled, err := json.Marshal(&b)
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}

	if got, want := string(marshalled), fmt.Sprintf(`"%s"`, revStr); got != want {
		t.Errorf("b.MarshalJSON() = %q want %q", got, want)
	}
}

func TestReversedBytes20UnmarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.ReversedBytes20
	err := json.Unmarshal([]byte(marshalled), &b)
	if err != nil {
		t.Fatalf("json.Unmarshal(): err: %s", err)
	}

	if got, want := b.String(), str; got != want {
		t.Errorf("b.UnmarshalJSON() = %q want %q", got, want)
	}
}

func TestReversedBytes20UnmarshalJSON_invalidStr(t *testing.T) {
	marshalled, err := json.Marshal([]string{"test"})
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}
	var b types.ReversedBytes20
	err = json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestReversedBytes20UnmarshalJSON_invalidHex(t *testing.T) {
	str := "1234o67890123456789012345678901234567890"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.ReversedBytes20
	err := json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestReversedBytes20Reverse(t *testing.T) {
	str := "1234567890123456789012345678901234567890"
	buf, _ := hex.DecodeString(str)
	var b types.ReversedBytes20
	copy(b[:], buf)
	var rev types.Bytes20
	b.Reverse(&rev)

	for i := range rev {
		if got, want := rev[i], b[len(b)-i-1]; got != want {
			t.Errorf("rev[%d] = %x want %x", i, got, want)
		}
	}
}

func TestNewBytes32FromString(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	b, err := types.NewBytes32FromString(str)
	if err != nil {
		t.Fatalf("NewBytes32FromString(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestNewBytes32FromString_invalidHex(t *testing.T) {
	if _, err := types.NewBytes32FromString("$234567890123456789012345678901234567890123456789012345678901234"); err == nil {
		t.Error("NewBytes32FromString(): err = nil want Error")
	}
}

func TestNewBytes32FromBytes(t *testing.T) {
	slice := make([]byte, types.Bytes32Size)
	for i := 0; i < types.Bytes32Size; i++ {
		slice[i] = byte(i)
	}

	bytes32 := types.NewBytes32FromBytes(slice)
	for i := 0; i < types.Bytes32Size; i++ {
		if bytes32[i] != byte(i) {
			t.Error("Invalid byte")
		}
	}
}

func TestNewBytes32FromBytes_emptySlice(t *testing.T) {
	bytes32 := types.NewBytes32FromBytes(nil)
	for i := 0; i < types.Bytes32Size; i++ {
		if bytes32[i] != byte(0) {
			t.Error("Invalid byte")
		}
	}
}

func TestNewBytes32FromBytes_bigSlice(t *testing.T) {
	slice := make([]byte, 2*types.Bytes32Size)
	for i := 0; i < 2*types.Bytes32Size; i++ {
		slice[i] = byte(i)
	}

	bytes32 := types.NewBytes32FromBytes(slice)
	for i := 0; i < types.Bytes32Size; i++ {
		if bytes32[i] != byte(i) {
			t.Error("Invalid byte")
		}
	}
}

func TestBytes32Compare(t *testing.T) {
	var nilBytes32 *types.Bytes32
	zero := &types.Bytes32{}
	nonZero, _ := types.NewBytes32FromString("1234567890123456789012345678901234567890123456789012345678901234")

	t.Run("Nil slice is equal to zero", func(t *testing.T) {
		assert.True(t, zero.EqualsBytes(nil))
	})

	t.Run("Empty slice is equal to zero", func(t *testing.T) {
		assert.True(t, zero.EqualsBytes([]byte{}))
	})

	t.Run("Nil bytes32 is not zero", func(t *testing.T) {
		assert.False(t, nilBytes32.Zero())
	})

	t.Run("Nil bytes32 comparison to zero", func(t *testing.T) {
		assert.NotEqual(t, 0, nilBytes32.Compare(zero))
	})

	t.Run("Nil bytes32 comparison", func(t *testing.T) {
		assert.NotEqual(t, 0, nilBytes32.Compare(nonZero))
	})
}

func TestBytes32String(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	buf, _ := hex.DecodeString(str)
	var b types.Bytes32
	copy(b[:], buf)

	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestBytes32Unstring(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	var b types.Bytes32
	if err := b.Unstring(str); err != nil {
		t.Fatalf("b.Unstring(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestBytes32Unstring_invalidHex(t *testing.T) {
	var b types.Bytes32
	if err := b.Unstring("123y567890123456789012345678901234567890123456789012345678901234"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestBytes32Unstring_invalidSize(t *testing.T) {
	var b types.Bytes32
	if err := b.Unstring("17890123456789012345678901234567890123456789012345"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestBytes32MarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	buf, _ := hex.DecodeString(str)
	var b types.Bytes32
	copy(b[:], buf)
	marshalled, err := json.Marshal(&b)
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}

	if got, want := string(marshalled), fmt.Sprintf(`"%s"`, str); got != want {
		t.Errorf("b.MarshalJSON() = %q want %q", got, want)
	}
}

func TestBytes32UnmarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.Bytes32
	err := json.Unmarshal([]byte(marshalled), &b)
	if err != nil {
		t.Fatalf("json.Unmarshal(): err: %s", err)
	}

	if got, want := b.String(), str; got != want {
		t.Errorf("b.UnmarshalJSON() = %q want %q", got, want)
	}
}

func TestBytes32UnmarshalJSON_invalidStr(t *testing.T) {
	marshalled, err := json.Marshal([]string{"test"})
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}
	var b types.Bytes32
	err = json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestBytes32UnmarshalJSON_invalidHex(t *testing.T) {
	str := "t234567890123456789012345678901234567890123456789012345678901234"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.Bytes32
	err := json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestBytes32Reverse(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	buf, _ := hex.DecodeString(str)
	var b types.Bytes32
	copy(b[:], buf)
	var rev types.ReversedBytes32
	b.Reverse(&rev)

	for i := range rev {
		if got, want := rev[i], b[len(b)-i-1]; got != want {
			t.Errorf("rev[%d] = %x want %x", i, got, want)
		}
	}
}

func TestNewReversedBytes32FromString(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	b, err := types.NewReversedBytes32FromString(str)
	if err != nil {
		t.Fatalf("NewReversedBytes32FromString(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestNewReversedBytes32FromString_invalidHex(t *testing.T) {
	if _, err := types.NewReversedBytes32FromString("^234567890123456789012345678901234567890123456789012345678901234"); err == nil {
		t.Error("NewReversedBytes32FromString(): err = nil want Error")
	}
}

func TestReversedBytes32String(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	revStr := "3412907856341290785634129078563412907856341290785634129078563412"
	buf, _ := hex.DecodeString(str)
	var b types.ReversedBytes32
	copy(b[:], buf)

	if got, want := b.String(), revStr; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestReversedBytes32Unstring(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	var b types.ReversedBytes32
	if err := b.Unstring(str); err != nil {
		t.Fatalf("b.Unstring(): err: %s", err)
	}
	if got, want := b.String(), str; got != want {
		t.Errorf("b.String() = %q want %q", got, want)
	}
}

func TestReversedBytes32Unstring_invalidHex(t *testing.T) {
	var b types.ReversedBytes32
	if err := b.Unstring("12345678901à3456789012345678901234567890123456789012345678901234"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestReversedBytes32Unstring_invalidSize(t *testing.T) {
	var b types.ReversedBytes32
	if err := b.Unstring("123456789015678456789012345678901234567890123456789012345678901234"); err == nil {
		t.Error("b.Unstring(): err = nil want Error")
	}
}

func TestReversedBytes32MarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	revStr := "3412907856341290785634129078563412907856341290785634129078563412"
	buf, _ := hex.DecodeString(str)
	var b types.ReversedBytes32
	copy(b[:], buf)
	marshalled, err := json.Marshal(&b)
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}

	if got, want := string(marshalled), fmt.Sprintf(`"%s"`, revStr); got != want {
		t.Errorf("b.MarshalJSON() = %q want %q", got, want)
	}
}

func TestReversedBytes32UnmarshalJSON(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.ReversedBytes32
	err := json.Unmarshal([]byte(marshalled), &b)
	if err != nil {
		t.Fatalf("json.Unmarshal(): err: %s", err)
	}

	if got, want := b.String(), str; got != want {
		t.Errorf("b.UnmarshalJSON() = %q want %q", got, want)
	}
}

func TestReversedBytes32UnmarshalJSON_invalidStr(t *testing.T) {
	marshalled, err := json.Marshal([]string{"test"})
	if err != nil {
		t.Fatalf("json.Marshal(): err: %s", err)
	}
	var b types.ReversedBytes32
	err = json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestReversedBytes32UnmarshalJSON_invalidHex(t *testing.T) {
	str := "12345'7890123456789012345678901234567890123456789012345678901234"
	marshalled := fmt.Sprintf(`"%s"`, str)
	var b types.ReversedBytes32
	err := json.Unmarshal([]byte(marshalled), &b)
	if err == nil {
		t.Error("json.Unmarshal(): err = nil want Error")
	}
}

func TestReversedBytes32Reverse(t *testing.T) {
	str := "1234567890123456789012345678901234567890123456789012345678901234"
	buf, _ := hex.DecodeString(str)
	var b types.ReversedBytes32
	copy(b[:], buf)
	var rev types.Bytes32
	b.Reverse(&rev)

	for i := range rev {
		if got, want := rev[i], b[len(b)-i-1]; got != want {
			t.Errorf("rev[%d] = %x want %x", i, got, want)
		}
	}
}
