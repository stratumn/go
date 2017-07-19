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
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

// Basic implements basic web socket server meant to be used in conjuction with
// an HTTP server.
type Basic struct {
	*Hub
	config        *BasicConfig
	bufConnConfig *BufferedConnConfig
	upgradeHandle UpgradeHandle
	msgAllocator  BasicMsgAllocator
	connChans     []chan *BufferedConn
	msgChans      []chan BasicConnMsg
}

// BasicConfig contains options for a basic web socket server.
type BasicConfig struct {
	ReadBufferSize  int               // Size of the read buffer in bytes
	WriteBufferSize int               // Size of the write buffer in bytes
	UpgradeHandle   UpgradeHandle     // Optional custom HTTP request upgrader
	MsgAllocator    BasicMsgAllocator // Optional custom message allocator
}

// BasicConnMsg contains a connection and a message read from that connection.
type BasicConnMsg struct {
	Conn *BufferedConn
	Msg  interface{}
}

// BasicMsgAllocator is a function that must initialize a message type before it
// is unmarshallied. A custom BasicMsgAllocator should be used when reading
// JSON messages so that they can be unmarshalled to an appropriate type. for
// instance:
//
//	func MyMsgAllocator(msg *interface{}) {
//		*msg = MyCustomMessageType{}
//	}
type BasicMsgAllocator func(*interface{})

// DefaultBasicMsgAllocator is the default function that allocates a message type
// before unmarshalling it. It allocate a map of strings to interface{}.
func DefaultBasicMsgAllocator(msg *interface{}) {
	*msg = map[string]interface{}{}
}

// NewBasic creates a new basic web socket server.
func NewBasic(config *BasicConfig, bufConnConfig *BufferedConnConfig) *Basic {
	var (
		handle UpgradeHandle
		alloc  BasicMsgAllocator
	)

	if config.UpgradeHandle != nil {
		handle = config.UpgradeHandle
	} else {
		// Use Gorilla web socket upgrader.
		upgrader := websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
		}
		handle = func(w http.ResponseWriter, r *http.Request, h http.Header) (PingableConn, error) {
			conn, err := upgrader.Upgrade(w, r, h)
			if err != nil {
				return nil, err
			}
			return GorrilaConn{conn}, nil
		}
	}

	if config.MsgAllocator != nil {
		alloc = config.MsgAllocator
	} else {
		alloc = DefaultBasicMsgAllocator
	}

	return &Basic{
		Hub:           NewHub(),
		config:        config,
		bufConnConfig: bufConnConfig,
		upgradeHandle: handle,
		msgAllocator:  alloc,
	}
}

// AddConnChannel adds a channel that will be sent new connections.
func (s *Basic) AddConnChannel(c chan *BufferedConn) {
	s.connChans = append(s.connChans, c)
}

// AddMsgChannel adds a channel that will be sent messages received by
// connections.
func (s *Basic) AddMsgChannel(c chan BasicConnMsg) {
	s.msgChans = append(s.msgChans, c)
}

// Handle handles an HTTP request for a web socket connection. The web socket
// route of the HTTP server should pass the writer and request to this function.
func (s *Basic) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgradeHandle(w, r, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"request": r,
		}).Warn("Failed to upgrade request to web socket connection")
		return
	}

	bufConn := NewBufferedConn(conn, s.bufConnConfig)

	for _, c := range s.connChans {
		c <- bufConn
	}

	s.Register(bufConn)

	errChan := make(chan error)

	go func() {
		errChan <- bufConn.Start()
	}()

	log.WithFields(log.Fields{
		"request":    r,
		"connection": bufConn,
	}).Info("Listening to web socket connection")

	for {
		connMsg := BasicConnMsg{Conn: bufConn}
		s.msgAllocator(&connMsg.Msg)

		if err = bufConn.ReadJSON(&connMsg.Msg); err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"request":    r,
				"connection": bufConn,
			}).Info("Closing web socket connection")
			break
		}

		for _, c := range s.msgChans {
			c <- connMsg
		}
	}

	s.Unregister(bufConn)

	if err = bufConn.Close(); err != nil {
		log.WithFields(log.Fields{
			"connection": bufConn,
			"error":      err,
		}).Warn("Failed to close web socket connection")
	}

	if err = <-errChan; err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"request":    r,
			"connection": bufConn,
		}).Warn("Web socket connection failed")
	}
}
