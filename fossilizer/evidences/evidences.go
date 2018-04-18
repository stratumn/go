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

// Package evidences does a blank import on all fossilizer evidences.
// It registers (de)serialization hooks for these evidence types.
// This package should only be imported in your main file when you
// will have to handle fossilized evidence from potentially any fossilizer.
package evidences

import (
	// Blank import to register fossilizer concrete evidence types.
	_ "github.com/stratumn/go-indigocore/batchfossilizer/evidences"
	_ "github.com/stratumn/go-indigocore/bcbatchfossilizer/evidences"
	_ "github.com/stratumn/go-indigocore/dummyfossilizer/evidences"
)
