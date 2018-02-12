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

package storehttp

import (
	"fmt"

	"github.com/stratumn/go-indigocore/jsonhttp"
	"github.com/stratumn/go-indigocore/store"
)

func newErrOffset(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "offset must be a positive integer"
	}
	return jsonhttp.NewErrBadRequest(msg)
}

func newErrLimit(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = fmt.Sprintf("limit must be a posive integer less than or equal to %d", store.MaxLimit)
	}
	return jsonhttp.NewErrBadRequest(msg)
}

func newErrPrevLinkHash(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "prevLinkHash must be a 64 byte long hexadecimal string"
	}
	return jsonhttp.NewErrBadRequest(msg)
}

func newErrLinkHashes(msg string) jsonhttp.ErrHTTP {
	if msg == "" {
		msg = "linkHashes must be an array of 64 byte long hexadecimal string"
	}
	return jsonhttp.NewErrBadRequest(msg)
}
