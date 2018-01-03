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

package main

import (
	"context"
	"flag"
	"os"

	"github.com/stratumn/sdk/fossilizer/fossilizerhttp"
	"github.com/stratumn/sdk/utils"

	"github.com/stratumn/sdk/bcbatchfossilizer"
	"github.com/stratumn/sdk/blockchain/btc/blockcypher"
	"github.com/stratumn/sdk/blockchain/btc/btctimestamper"
)

var (
	key = flag.String("wif", os.Getenv("BTCFOSSILIZER_WIF"), "wallet import format key")

	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	fossilizerhttp.RegisterFlags()
	blockcypher.RegisterFlags()
	btctimestamper.RegisterFlags()
	bcbatchfossilizer.RegisterFlags()
}

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx = utils.CancelOnInterrupt(ctx)

	bcy := blockcypher.RunWithFlags(ctx, *key)
	ts := btctimestamper.InitializeWithFlags(version, commit, *key, bcy, bcy)
	a := bcbatchfossilizer.RunWithFlags(ctx, version, commit, ts)
	fossilizerhttp.RunWithFlags(ctx, a)
}
