// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license that can be found in the
// LICENSE file.

package batchfossilizer

import (
	"encoding/gob"
	"os"

	"github.com/stratumn/sdk/types"
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
