// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"encoding/gob"

	"github.com/stratumn/go/types"
)

type fossil struct {
	Data types.Bytes32
	Meta []byte
}

func newFossilFromDecoder(dec *gob.Decoder) (f *fossil, err error) {
	err = dec.Decode(&f)
	return
}

func (f *fossil) write(enc *gob.Encoder) error {
	return enc.Encode(f)
}
