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

import log "github.com/Sirupsen/logrus"

// Hub manages a list of web socket connections.
// Messages can be broadcasted in JSON form to the list of connections.
// Connections may also be tagged, and messages can be broadcasted only to
// connections that have a certain tag.
type Hub struct {
	// conns keeps a set of tags for each connection.
	conns map[Writer]map[interface{}]struct{}
	// tags keeps a set of connections for each tag.
	tags map[interface{}]map[Writer]struct{}
	// We use channels within a select for operations since maps and the
	// underlying web socket implementation are not concurrently safe.
	stopChan      chan struct{}
	regChan       chan Writer
	unregChan     chan Writer
	tagChan       chan connTag
	untagChan     chan connTag
	broadcastChan chan msgTag
}

// Used by the tag channel.
type connTag struct {
	conn Writer
	tag  interface{}
}

// Used by the broadcast channel.
type msgTag struct {
	msg interface{}
	tag interface{}
}

// NewHub creates a new hub.
func NewHub() *Hub {
	return &Hub{
		map[Writer]map[interface{}]struct{}{},
		map[interface{}]map[Writer]struct{}{},
		make(chan struct{}),
		make(chan Writer),
		make(chan Writer),
		make(chan connTag),
		make(chan connTag),
		make(chan msgTag),
	}
}

// Start starts managing the client connections.
func (h *Hub) Start() {
	for {
		select {
		case <-h.stopChan:
			return
		case c := <-h.regChan:
			h.conns[c] = map[interface{}]struct{}{}
		case c := <-h.unregChan:
			// Remove connection from tags.
			for t := range h.conns[c] {
				delete(h.tags[t], c)
			}
			delete(h.conns, c)
		case t := <-h.tagChan:
			// Add tag to connection.
			h.conns[t.conn][t.tag] = struct{}{}
			// Add connection to tag.
			if _, ok := h.tags[t.tag]; !ok {
				h.tags[t.tag] = map[Writer]struct{}{}
			}
			h.tags[t.tag][t.conn] = struct{}{}
		case t := <-h.untagChan:
			delete(h.conns[t.conn], t.tag)
			delete(h.tags[t.tag], t.conn)
		case m := <-h.broadcastChan:
			if m.tag == nil {
				for c := range h.conns {
					writeMsg(c, m.msg)
				}
			} else {
				for c := range h.tags[m.tag] {
					writeMsg(c, m.msg)
				}
			}
		}
	}
}

// Stop stops managing the client connections.
func (h *Hub) Stop() {
	h.stopChan <- struct{}{}
}

// Register adds a connection to the list.
func (h *Hub) Register(conn Writer) {
	h.regChan <- conn
}

// Unregister removes a connection from the list.
func (h *Hub) Unregister(conn Writer) {
	h.unregChan <- conn
}

// Tag adds a tag to a connection.
func (h *Hub) Tag(conn Writer, tag interface{}) {
	h.tagChan <- connTag{conn, tag}
}

// Untag remotes a tag from a connection.
func (h *Hub) Untag(conn Writer, tag interface{}) {
	h.untagChan <- connTag{conn, tag}
}

// Broadcast broadcasts the JSON representation of a message. If tag is nil,
// it broadcasts the message to every connection. Otherwise it broadcasts the
// message only to connections that have that tag.
func (h *Hub) Broadcast(msg interface{}, tag interface{}) {
	h.broadcastChan <- msgTag{msg, tag}
}

// Writes a message to a connection and logs errors.
func writeMsg(conn Writer, msg interface{}) {
	if err := conn.WriteJSON(msg); err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"connection": conn,
			"message":    msg,
		}).Warn("Failed to broadcast message to client")
	}
}
