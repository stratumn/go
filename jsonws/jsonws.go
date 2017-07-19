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

// Package jsonws defines functionality to deal with web sockets and JSON.
package jsonws

import "net/http"

// UpgradeHandle is a function that upgrades an HTTP connection to a web socket
// connection.
type UpgradeHandle func(w http.ResponseWriter, r *http.Request, h http.Header) (PingableConn, error)
