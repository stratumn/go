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
	"os"

	"github.com/stratumn/go-core/batchfossilizer"
	"github.com/stratumn/go-core/blockchain/btc"
	"github.com/stratumn/go-core/blockchain/btc/blockcypher"
	"github.com/stratumn/go-core/blockchain/btc/btctimestamper"
	"github.com/stratumn/go-core/blockchainfossilizer"
	"github.com/stratumn/go-core/cloud/aws"
	"github.com/stratumn/go-core/fossilizer"
	"github.com/stratumn/go-core/fossilizer/fossilizerhttp"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/util"
)

var (
	key       = flag.String("wif", os.Getenv("BTCFOSSILIZER_WIF"), "wallet import format key")
	bcyAPIKey = flag.String("bcyapikey", "", "BlockCypher API key")
	fee       = flag.Int64("fee", int64(15000), "transaction fee (satoshis)")

	queueType    = flag.String("queuetype", AWSQueue, "queue implementation ('dummy' or 'aws')")
	fossilsQueue = flag.String("fossilsqueue", DefaultFossilsQueue, "name of the pending fossils queue")

	exporter      = flag.String("exporter", NoExporter, "exporter for fossilizer events ('', 'console' or 'aws')")
	exporterQueue = flag.String("exporterqueue", DefaultExporterQueue, "name of the fossilizer events queue")

	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	fossilizerhttp.RegisterFlags()
	batchfossilizer.RegisterFlags()
	monitoring.RegisterFlags()
	aws.RegisterFlags()

	monitoring.SetVersion(version, commit)
}

func main() {
	flag.Parse()

	ctx := util.CancelOnInterrupt(context.Background())

	if *key == "" {
		monitoring.LogEntry().Fatal("A WIF encoded private key is required")
	}

	network, err := btc.GetNetworkFromWIF(*key)
	if err != nil {
		monitoring.LogEntry().WithField("error", err).Fatal()
	}

	bcy := blockcypher.New(&blockcypher.Config{
		Network: network,
		APIKey:  *bcyAPIKey,
	})

	ts, err := btctimestamper.New(&btctimestamper.Config{
		Fee:           *fee,
		WIF:           *key,
		Broadcaster:   bcy,
		UnspentFinder: bcy,
	})
	if err != nil {
		monitoring.LogEntry().WithField("error", err).Fatal()
	}

	a := monitoring.NewFossilizerAdapter(
		batchfossilizer.New(ctx,
			batchfossilizer.ConfigFromFlags(version, commit),
			blockchainfossilizer.New(&blockchainfossilizer.Config{
				Commit:      commit,
				Version:     version,
				Timestamper: ts,
			}),
			QueueFromFlags(),
		),
		"btcfossilizer",
	)

	// Run an exporter for fossilizer events.
	exporter := ExporterFromFlags()
	if exporter != nil {
		go fossilizer.RunExporter(ctx, a, exporter)
	}

	fossilizerhttp.RunWithFlags(ctx, a)
}
