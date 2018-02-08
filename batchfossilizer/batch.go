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

package batchfossilizer

import (
	"encoding/gob"
	"os"

	"github.com/stratumn/go-indigocore/types"
)

type batch struct {
	data    []types.Bytes32
	meta    [][]byte
	file    *os.File
	encoder *gob.Encoder
}

func newBatch(maxLeaves int) *batch {
	return &batch{
		data: make([]types.Bytes32, 0, maxLeaves),
		meta: make([][]byte, 0, maxLeaves),
	}
}

func (b *batch) append(f *fossil) {
	b.data = append(b.data, f.Data)
	b.meta = append(b.meta, f.Meta)
}

func (b *batch) open(path string) (err error) {
	flags := os.O_APPEND | os.O_WRONLY | os.O_EXCL | os.O_CREATE
	if b.file, err = os.OpenFile(path, flags, FilePerm); err != nil {
		return
	}
	b.encoder = gob.NewEncoder(b.file)
	return
}

func (b *batch) close() (err error) {
	if b.file != nil {
		err = b.file.Close()
		b.file = nil
	}
	return
}
