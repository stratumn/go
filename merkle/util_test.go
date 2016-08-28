// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package merkle_test

import (
	"encoding/json"
	"io/ioutil"

	"github.com/stratumn/goprivate/merkle"
)

func loadPath(filename string, path *merkle.Path) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, path); err != nil {
		return err
	}
	return nil
}
