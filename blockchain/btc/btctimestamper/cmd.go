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

package btctimestamper

import (
	"flag"

	"github.com/stratumn/sdk/blockchain/btc"

	log "github.com/sirupsen/logrus"
)

var (
	fee int64
)

// RegisterFlags registers the flags used by InitializeWithFlags.
func RegisterFlags() {
	flag.Int64Var(&fee, "fee", DefaultFee, "transaction fee (satoshis)")
}

// InitializeWithFlags should be called after RegisterFlags and flag.Parse to initialize
// a bcbatchfossilizer using flag values.
func InitializeWithFlags(version, commit string, key string, unspentFinder btc.UnspentFinder, broadcaster btc.Broadcaster) *Timestamper {
	ts, err := New(&Config{
		UnspentFinder: unspentFinder,
		Broadcaster:   broadcaster,
		WIF:           key,
		Fee:           fee,
	})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create Bitcoin timestamper")
	}
	return ts
}
