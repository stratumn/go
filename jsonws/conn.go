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

import "time"
import "github.com/gorilla/websocket"

// Conn must be implemented by a web socket connection.
type Conn interface {
	Writer

	// Closes the connection.
	Close() error

	// Reads JSON from the connection. It blocks until a value is received.
	ReadJSON(v interface{}) error

	// SetReadLimit sets the maximum size for a message read.
	SetReadLimit(limit int64)

	// SetReadDeadline sets the read deadline.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the write deadline.
	SetWriteDeadline(t time.Time) error

	// SetPongHandler sets the handler for pong messages received.
	SetPongHandler(h func(appData string) error)
}

// Writer must be able to write JSON messages.
type Writer interface {
	// Writes JSON to the connection.
	WriteJSON(v interface{}) error
}

// PingableConn must be able to send a ping control message.
type PingableConn interface {
	Conn

	// Sends a ping control message.
	Ping() error
}

// GorrilaConn implements github.com/stratumn/sdk/jsonws/Conn using a Gorrila
// web socket connection.
type GorrilaConn struct {
	*websocket.Conn
}

// Ping implements github.com/stratumn/sdk/jsonws/Conn.Ping.
func (c GorrilaConn) Ping() error {
	return c.WriteMessage(websocket.PingMessage, []byte{})
}
