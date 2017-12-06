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

package tmpoptestcases

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/tmpop"
)

func getTestFile(t *testing.T) string {
	_, currentFilename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("Could not get information about the current context")
	}
	testFile := filepath.Join(filepath.Dir(currentFilename), "testdata/rules.json")
	return testFile
}

// TestValidation tests what happens when validating a segment from a json-schema based validation file
func (f Factory) TestValidation(t *testing.T) {
	testFilename := getTestFile(t)
	h, req := f.newTMPop(t, &tmpop.Config{ValidatorFilename: testFilename})
	defer f.free()

	h.BeginBlock(req)

	l := cstesting.RandomLinkWithProcess("testProcess")
	l.Meta["action"] = "init"
	l.State["string"] = "test"
	tx := makeCreateLinkTx(t, l)
	res := h.DeliverTx(tx)

	if res.IsErr() {
		t.Errorf("a.DeliverTx(): failed: %v", res.Log)
	}

	l = cstesting.RandomLinkWithProcess("testProcess")
	l.Meta["action"] = "init"
	l.State["string"] = 42
	tx = makeCreateLinkTx(t, l)
	res = h.DeliverTx(tx)

	if !res.IsErr() {
		t.Error("a.DeliverTx(): want error")
	}
	if res.Code != tmpop.CodeTypeValidation {
		t.Errorf("res.Code = got %d want %d", res.Code, tmpop.CodeTypeValidation)
	}
}
