// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package bcbatchfossilizer

import (
	"testing"

	"github.com/stratumn/goprivate/batchfossilizer"
	"github.com/stratumn/goprivate/blockchain/dummytimestamper"
)

func BenchmarkFossilize_MaxLeaves100(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval:  interval,
		MaxLeaves: 100,
	})
}

func BenchmarkFossilize_MaxLeaves1000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval:  interval,
		MaxLeaves: 1000,
	})
}

func BenchmarkFossilize_MaxLeaves10000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval:  interval,
		MaxLeaves: 10000,
	})
}

func BenchmarkFossilize_MaxLeaves100000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval:  interval,
		MaxLeaves: 100000,
	})
}

func BenchmarkFossilize_MaxLeaves1000000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		Interval:  interval,
		MaxLeaves: 1000000,
	})
}
