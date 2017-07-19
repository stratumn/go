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

import "time"

// MockConn is used to mock a connection.
//
// It implements github.com/stratumn/sdk/jsonws.Conn and
// github.com/stratumn/sdk/jsonws.PingableConn.
type MockConn struct {
	// The mock for the Close function.
	MockClose MockClose

	// The mock for the WriteJSON function.
	MockWriteJSON MockWriteJSON

	// The mock for the ReadJSON function.
	MockReadJSON MockReadJSON

	// The mock for the SetReadLimit function.
	MockSetReadLimit MockSetReadLimit

	// The mock for the SetReadDeadline function.
	MockSetReadDeadline MockSetReadDeadline

	// The mock for the SetWriteDeadline function.
	MockSetWriteDeadline MockSetWriteDeadline

	// The mock for the SetPongHandler function.
	MockSetPongHandler MockSetPongHandler

	// The mock for the Ping function.
	MockPing MockPing
}

// MockClose mocks the Close function.
type MockClose struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() error
}

// MockWriteJSON mocks the WriteJSON function.
type MockWriteJSON struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []interface{}

	// The last channel that was passed.
	LastCalledWith interface{}

	// An optional implementation of the function.
	Fn func(interface{}) error
}

// MockReadJSON mocks the ReadJSON function.
type MockReadJSON struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []interface{}

	// The last channel that was passed.
	LastCalledWith interface{}

	// An optional implementation of the function.
	Fn func(interface{}) error
}

// MockSetReadLimit mocks the SetReadLimit function.
type MockSetReadLimit struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []int64

	// The last channel that was passed.
	LastCalledWith int64

	// An optional implementation of the function.
	Fn func(int64)
}

// MockSetReadDeadline mocks the SetReadDeadline function.
type MockSetReadDeadline struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []time.Time

	// The last channel that was passed.
	LastCalledWith time.Time

	// An optional implementation of the function.
	Fn func(time.Time) error
}

// MockSetWriteDeadline mocks the SetWriteDeadline function.
type MockSetWriteDeadline struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []time.Time

	// The last channel that was passed.
	LastCalledWith time.Time

	// An optional implementation of the function.
	Fn func(time.Time) error
}

// MockSetPongHandler mocks the SetPongHandler function.
type MockSetPongHandler struct {
	// The number of times the function was called.
	CalledCount int

	// The channel that was passed to each call.
	CalledWith []func(string) error

	// The last channel that was passed.
	LastCalledWith func(string) error

	// An optional implementation of the function.
	Fn func(func(string) error)
}

// MockPing mocks the SetPing function.
type MockPing struct {
	// The number of times the function was called.
	CalledCount int

	// An optional implementation of the function.
	Fn func() error
}

// Close implements
// github.com/stratumn/sdk/jsonws.Conn.Close.
func (a *MockConn) Close() error {
	a.MockClose.CalledCount++

	if a.MockClose.Fn != nil {
		return a.MockClose.Fn()
	}

	return nil
}

// WriteJSON implements
// github.com/stratumn/sdk/jsonws.Conn.WriteJSON.
func (a *MockConn) WriteJSON(msg interface{}) error {
	a.MockWriteJSON.CalledCount++
	a.MockWriteJSON.CalledWith = append(a.MockWriteJSON.CalledWith, msg)
	a.MockWriteJSON.LastCalledWith = msg

	if a.MockWriteJSON.Fn != nil {
		return a.MockWriteJSON.Fn(msg)
	}

	return nil
}

// ReadJSON implements
// github.com/stratumn/sdk/jsonws.Conn.ReadJSON.
func (a *MockConn) ReadJSON(msg interface{}) error {
	a.MockReadJSON.CalledCount++
	a.MockReadJSON.CalledWith = append(a.MockReadJSON.CalledWith, msg)
	a.MockReadJSON.LastCalledWith = msg

	if a.MockReadJSON.Fn != nil {
		return a.MockReadJSON.Fn(msg)
	}

	return nil
}

// SetReadLimit implements
// github.com/stratumn/sdk/jsonws.Conn.SetReadLimit.
func (a *MockConn) SetReadLimit(limit int64) {
	a.MockSetReadLimit.CalledCount++
	a.MockSetReadLimit.CalledWith = append(a.MockSetReadLimit.CalledWith, limit)
	a.MockSetReadLimit.LastCalledWith = limit

	if a.MockSetReadLimit.Fn != nil {
		a.MockSetReadLimit.Fn(limit)
	}
}

// SetReadDeadline implements
// github.com/stratumn/sdk/jsonws.Conn.SetReadDeadline.
func (a *MockConn) SetReadDeadline(t time.Time) error {
	a.MockSetReadDeadline.CalledCount++
	a.MockSetReadDeadline.CalledWith = append(a.MockSetReadDeadline.CalledWith, t)
	a.MockSetReadDeadline.LastCalledWith = t

	if a.MockSetReadDeadline.Fn != nil {
		return a.MockSetReadDeadline.Fn(t)
	}

	return nil
}

// SetWriteDeadline implements
// github.com/stratumn/sdk/jsonws.Conn.SetWriteDeadline.
func (a *MockConn) SetWriteDeadline(t time.Time) error {
	a.MockSetWriteDeadline.CalledCount++
	a.MockSetWriteDeadline.CalledWith = append(a.MockSetWriteDeadline.CalledWith, t)
	a.MockSetWriteDeadline.LastCalledWith = t

	if a.MockSetWriteDeadline.Fn != nil {
		return a.MockSetWriteDeadline.Fn(t)
	}

	return nil
}

// SetPongHandler implements
// github.com/stratumn/sdk/jsonws.Conn.SetPongHandler.
func (a *MockConn) SetPongHandler(h func(string) error) {
	a.MockSetPongHandler.CalledCount++
	a.MockSetPongHandler.CalledWith = append(a.MockSetPongHandler.CalledWith, h)
	a.MockSetPongHandler.LastCalledWith = h

	if a.MockSetPongHandler.Fn != nil {
		a.MockSetPongHandler.Fn(h)
	}
}

// Ping implements github.com/stratumn/sdk/jsonws.PingableConn.Ping.
func (a *MockConn) Ping() error {
	a.MockPing.CalledCount++

	if a.MockPing.Fn != nil {
		return a.MockPing.Fn()
	}

	return nil
}
