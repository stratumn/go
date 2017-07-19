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

package jsonws

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stratumn/sdk/jsonws/jsonwstesting"
)

func TestHubBroadcast(t *testing.T) {
	h := NewHub()
	go h.Start()

	c1 := &jsonwstesting.MockConn{}
	c2 := &jsonwstesting.MockConn{}

	// Check errors don't stop the hub.
	c1.MockWriteJSON.Fn = func(interface{}) error {
		return errors.New("test")
	}

	h.Register(c1)
	h.Register(c2)

	m := map[string]string{"msg": "hello"}

	h.Broadcast(m, nil)
	h.Stop()

	if got, want := c1.MockWriteJSON.CalledCount, 1; got != want {
		t.Errorf(`c1.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}
	if got, want := c2.MockWriteJSON.CalledCount, 1; got != want {
		t.Errorf(`c2.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}

	if got, want := c1.MockWriteJSON.LastCalledWith, m; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("c1.MockWriteJSON.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
	if got, want := c2.MockWriteJSON.LastCalledWith, m; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("c2.MockWriteJSON.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestHubUnregister(t *testing.T) {
	h := NewHub()
	go h.Start()

	c1 := &jsonwstesting.MockConn{}
	c2 := &jsonwstesting.MockConn{}

	h.Register(c1)
	h.Register(c2)
	h.Unregister(c1)

	m := map[string]string{"msg": "hello"}

	h.Broadcast(m, nil)
	h.Stop()

	if got, want := c1.MockWriteJSON.CalledCount, 0; got != want {
		t.Errorf(`c1.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}
	if got, want := c2.MockWriteJSON.CalledCount, 1; got != want {
		t.Errorf(`c2.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}
}

func TestHubTag(t *testing.T) {
	h := NewHub()
	go h.Start()

	c1 := &jsonwstesting.MockConn{}
	c2 := &jsonwstesting.MockConn{}

	h.Register(c1)
	h.Register(c2)

	h.Tag(c1, "test")

	m := map[string]string{"msg": "hello"}

	h.Broadcast(m, "test")

	// Check Unregister removes connection from tag. This call to Broadcast
	// should not reach any connection.
	h.Unregister(c1)
	h.Broadcast(m, "test")
	h.Stop()

	if got, want := c1.MockWriteJSON.CalledCount, 1; got != want {
		t.Errorf(`c1.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}
	if got, want := c2.MockWriteJSON.CalledCount, 0; got != want {
		t.Errorf(`c2.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}

	if got, want := c1.MockWriteJSON.LastCalledWith, m; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("c1.MockWriteJSON.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}

func TestHubUntag(t *testing.T) {
	h := NewHub()
	go h.Start()

	c1 := &jsonwstesting.MockConn{}
	c2 := &jsonwstesting.MockConn{}

	h.Register(c1)
	h.Register(c2)

	h.Tag(c1, "test")
	h.Tag(c2, "test")
	h.Untag(c2, "test")

	m := map[string]string{"msg": "hello"}

	h.Broadcast(m, "test")
	h.Stop()

	if got, want := c1.MockWriteJSON.CalledCount, 1; got != want {
		t.Errorf(`c1.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}
	if got, want := c2.MockWriteJSON.CalledCount, 0; got != want {
		t.Errorf(`c2.MockWriteJSON.CalledCount = %d want %d`, got, want)
	}

	if got, want := c1.MockWriteJSON.LastCalledWith, m; !reflect.DeepEqual(got, want) {
		gotJS, _ := json.MarshalIndent(got, "", "  ")
		wantJS, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("c1.MockWriteJSON.LastCalledWith = %s\n want %s", gotJS, wantJS)
	}
}
