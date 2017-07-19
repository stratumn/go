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
	"time"

	log "github.com/Sirupsen/logrus"
)

// BufferedConn wraps a connection so that writes are buffered and not blocking
// unless the channel is full. Is it a higher level type that also deals with
// control messages and timeouts. It requires an underlying PingableConn.
type BufferedConn struct {
	conn      PingableConn
	config    *BufferedConnConfig
	closeChan chan struct{}
	writeChan chan interface{}
}

// BufferedConnConfig contains options for a buffered connection.
type BufferedConnConfig struct {
	Size         int           // Size of the write channel
	WriteTimeout time.Duration // Time allowed to write a message
	PongTimeout  time.Duration // Time allowed to read next pong
	PingInterval time.Duration // Interval between two pings (< PongTimeout)
	MaxMsgSize   int64         // Maximum size of input message in bytes
}

// NewBufferedConn creates a new buffered connection from a pingable connection.
func NewBufferedConn(conn PingableConn, config *BufferedConnConfig) *BufferedConn {
	return &BufferedConn{
		conn,
		config,
		make(chan struct{}),
		make(chan interface{}, config.Size),
	}
}

// Start starts the buffered connection. It will stop with an error if a write
// failed. It will also ping the connection at regulart intervals.
func (c *BufferedConn) Start() error {
	// Configure control messages.
	c.conn.SetReadLimit(c.config.MaxMsgSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.PongTimeout)); err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"connection": c,
		}).Error("Failed to set read deadline")
	}
	c.conn.SetPongHandler(func(string) error {
		e := c.conn.SetReadDeadline(time.Now().Add(c.config.PongTimeout))
		if e != nil {
			log.WithFields(log.Fields{
				"error":      e,
				"connection": c,
			}).Error("Failed to set read deadline")
		}
		return e
	})

	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closeChan:
			return nil
		case v := <-c.writeChan:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
			if err := c.conn.WriteJSON(v); err != nil {
				return err
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
			if err := c.conn.Ping(); err != nil {
				return err
			}
		}
	}
}

// Close closes the connection.
func (c *BufferedConn) Close() error {
	c.closeChan <- struct{}{}
	return c.conn.Close()
}

// WriteJSON writes JSON to the connection.
func (c *BufferedConn) WriteJSON(v interface{}) error {
	c.writeChan <- v
	return nil
}

// ReadJSON reads JSON from the connection.It blocks until a value is received
func (c *BufferedConn) ReadJSON(v interface{}) error {
	return c.conn.ReadJSON(v)
}
