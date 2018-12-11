// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package main

import (
	"context"
	"flag"

	"github.com/stratumn/go-core/batchfossilizer"
	"github.com/stratumn/go-core/dummyfossilizer"
	"github.com/stratumn/go-core/fossilizer/dummyqueue"
	"github.com/stratumn/go-core/fossilizer/fossilizerhttp"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/util"
)

var (
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	fossilizerhttp.RegisterFlags()
	batchfossilizer.RegisterFlags()
	monitoring.RegisterFlags()
}

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx = util.CancelOnInterrupt(ctx)

	a := monitoring.NewFossilizerAdapter(
		batchfossilizer.New(ctx,
			batchfossilizer.ConfigFromFlags(version, commit),
			dummyfossilizer.New(&dummyfossilizer.Config{
				Version: version,
				Commit:  commit,
			}),
			dummyqueue.New(),
		),
		"dummybatchfossilizer",
	)
	fossilizerhttp.RunWithFlags(ctx, a)
}
