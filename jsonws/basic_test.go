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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stratumn/sdk/jsonws/jsonwstesting"
)

func testUpgradeHandle(w http.ResponseWriter, r *http.Request, h http.Header) (PingableConn, error) {
	conn := jsonwstesting.MockConn{}
	read := false
	conn.MockReadJSON.Fn = func(v interface{}) error {
		if read {
			return io.EOF
		}
		read = true
		m := v.(*interface{})
		*m = map[string]string{"msg": "test"}
		return nil
	}
	return &conn, nil
}

func testMsgAllocator(msg *interface{}) {
	*msg = map[string]string{}
}

func TestBasicAddConnChannel(t *testing.T) {
	ws := NewBasic(&BasicConfig{
		UpgradeHandle: testUpgradeHandle,
	}, &BufferedConnConfig{
		PingInterval: time.Second,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/ws", nil)
	c := make(chan *BufferedConn)
	ws.AddConnChannel(c)

	go ws.Handle(w, r)

	select {
	case got := <-c:
		if got == nil {
			t.Errorf("<-c = nil want not nil")
		}
	case <-time.After(time.Second):
		t.Errorf("no connection sent to channel")
	}
}

func TestBasicAddMsgChannel(t *testing.T) {
	ws := NewBasic(&BasicConfig{
		UpgradeHandle: testUpgradeHandle,
		MsgAllocator:  testMsgAllocator,
	}, &BufferedConnConfig{
		PingInterval: time.Second,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/ws", nil)
	c := make(chan BasicConnMsg)
	ws.AddMsgChannel(c)

	go ws.Start()
	go ws.Handle(w, r)
	defer ws.Stop()

	select {
	case msg := <-c:
		if got, want := msg.Msg.(map[string]string)["msg"], "test"; got != want {
			t.Errorf(`msg = %q want %q`, got, want)
		}
	case <-time.After(time.Second):
		t.Errorf("no message sent to channel")
	}
}
