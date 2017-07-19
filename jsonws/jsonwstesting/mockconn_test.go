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

package jsonwstesting

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestMockConn_WriteJSON(t *testing.T) {
	a := &MockConn{}
	m := map[string]string{"msg": "hello"}

	err := a.WriteJSON(m)
	if err != nil {
		t.Fatalf("a.WriteJSON(): err: %s", err)
	}

	a.MockWriteJSON.Fn = func(interface{}) error { return nil }
	err = a.WriteJSON(m)
	if err != nil {
		t.Fatalf("a.WriteJSON(): err: %s", err)
	}

	if got, want := a.MockWriteJSON.CalledCount, 2; got != want {
		t.Errorf(`a.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockWriteJSON.CalledWith, []interface{}{m, m}; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockWriteJSON.CalledWith = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockWriteJSON.LastCalledWith, m; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockWriteJSON.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestMockConn_ReadJSON(t *testing.T) {
	a := &MockConn{}
	m := map[string]string{"msg": "hello"}

	err := a.ReadJSON(m)
	if err != nil {
		t.Fatalf("a.ReadJSON(): err: %s", err)
	}

	a.MockReadJSON.Fn = func(interface{}) error { return nil }
	err = a.ReadJSON(m)
	if err != nil {
		t.Fatalf("a.ReadJSON(): err: %s", err)
	}

	if got, want := a.MockReadJSON.CalledCount, 2; got != want {
		t.Errorf(`a.MockReadJSON.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockReadJSON.CalledWith, []interface{}{m, m}; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockReadJSON.CalledWith = %s\n want %s", gotJS, wantJS)
	}
	if got, want := a.MockReadJSON.LastCalledWith, m; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("a.MockReadJSON.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestMockConn_Close(t *testing.T) {
	a := &MockConn{}

	err := a.Close()
	if err != nil {
		t.Fatalf("a.Close(): err: %s", err)
	}

	a.MockClose.Fn = func() error { return nil }
	err = a.Close()
	if err != nil {
		t.Fatalf("a.Close(): err: %s", err)
	}

	if got, want := a.MockClose.CalledCount, 2; got != want {
		t.Errorf(`a.MockClose.CalledCount = %d want %d`, got, want)
	}
}

func TestMockConn_SetReadLimit(t *testing.T) {
	a := &MockConn{}

	a.SetReadLimit(10)
	a.MockSetReadLimit.Fn = func(_ int64) {}
	a.SetReadLimit(20)

	if got, want := a.MockSetReadLimit.CalledCount, 2; got != want {
		t.Errorf(`a.MockSetReadLimit.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockSetReadLimit.CalledWith, []int64{10, 20}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSetReadLimit.CalledWith = %q want %q", got, want)
	}
	if got, want := a.MockSetReadLimit.LastCalledWith, int64(20); got != want {
		t.Errorf("a.MockSetReadLimit.LastCalledWith = %d want %d", got, want)
	}
}

func TestMockConn_SetReadDeadline(t *testing.T) {
	a := &MockConn{}
	v := time.Now()

	a.SetReadDeadline(v)
	a.MockSetReadDeadline.Fn = func(_ time.Time) error { return nil }
	a.SetReadDeadline(v)

	if got, want := a.MockSetReadDeadline.CalledCount, 2; got != want {
		t.Errorf(`a.MockSetReadDeadline.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockSetReadDeadline.CalledWith, []time.Time{v, v}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSetReadDeadline.CalledWith = %q want %q", got, want)
	}
	if got, want := a.MockSetReadDeadline.LastCalledWith, v; got != want {
		t.Errorf("a.MockSetReadDeadline.LastCalledWith = %d want %d", got, want)
	}
}

func TestMockConn_SetWriteDeadline(t *testing.T) {
	a := &MockConn{}
	v := time.Now()

	a.SetWriteDeadline(v)
	a.MockSetWriteDeadline.Fn = func(_ time.Time) error { return nil }
	a.SetWriteDeadline(v)

	if got, want := a.MockSetWriteDeadline.CalledCount, 2; got != want {
		t.Errorf(`a.MockSetWriteDeadline.CalledCount = %d want %d`, got, want)
	}
	if got, want := a.MockSetWriteDeadline.CalledWith, []time.Time{v, v}; !reflect.DeepEqual(got, want) {
		t.Errorf("a.MockSetWriteDeadline.CalledWith = %q want %q", got, want)
	}
	if got, want := a.MockSetWriteDeadline.LastCalledWith, v; got != want {
		t.Errorf("a.MockSetWriteDeadline.LastCalledWith = %d want %d", got, want)
	}
}

func TestMockConn_SetPongHandler(t *testing.T) {
	a := &MockConn{}
	f := func(_ string) error { return nil }

	a.SetPongHandler(f)
	a.MockSetPongHandler.Fn = func(_ func(string) error) {}
	a.SetPongHandler(f)

	if got, want := a.MockSetPongHandler.CalledCount, 2; got != want {
		t.Errorf(`a.MockSetPongHandler.CalledCount = %d want %d`, got, want)
	}
}

func TestMockConn_Ping(t *testing.T) {
	a := &MockConn{}

	a.Ping()
	a.MockPing.Fn = func() error { return nil }
	a.Ping()

	if got, want := a.MockPing.CalledCount, 2; got != want {
		t.Errorf(`a.MockPing.CalledCount = %d want %d`, got, want)
	}
}
