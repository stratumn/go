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

package bcbatchfossilizer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/sdk/batchfossilizer"
	"github.com/stratumn/sdk/blockchain/dummytimestamper"
)

func BenchmarkFossilize_MaxLeaves100(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 100,
	})
}

func BenchmarkFossilize_MaxLeaves1000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 1000,
	})
}

func BenchmarkFossilize_MaxLeaves10000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 10000,
	})
}

func BenchmarkFossilize_MaxLeaves100000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 100000,
	})
}

func BenchmarkFossilize_MaxLeaves1000000(b *testing.B) {
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 1000000,
	})
}

func BenchmarkFossilize_Path_MaxLeaves100(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 100,
		Path:      path,
	})
}

func BenchmarkFossilize_Path_MaxLeaves1000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 1000,
		Path:      path,
	})
}

func BenchmarkFossilize_Path_MaxLeaves10000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 10000,
		Path:      path,
	})
}

func BenchmarkFossilize_Path_MaxLeaves100000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 100000,
		Path:      path,
	})
}

func BenchmarkFossilize_Path_MaxLeaves1000000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 1000000,
		Path:      path,
	})
}

func BenchmarkFossilize_FSync_MaxLeaves100(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 100,
		Path:      path,
		FSync:     true,
	})
}

func BenchmarkFossilize_FSync_MaxLeaves1000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 1000,
		Path:      path,
		FSync:     true,
	})
}

func BenchmarkFossilize_FSync_MaxLeaves10000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 10000,
		Path:      path,
		FSync:     true,
	})
}

func BenchmarkFossilize_FSync_MaxLeaves100000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 100000,
		Path:      path,
		FSync:     true,
	})
}

func BenchmarkFossilize_FSync_MaxLeaves1000000(b *testing.B) {
	path, err := ioutil.TempDir("", "bcbatchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{
		HashTimestamper: dummytimestamper.Timestamper{},
	}, &batchfossilizer.Config{
		StopBatch: true,
		Interval:  testInterval,
		MaxLeaves: 1000000,
		Path:      path,
		FSync:     true,
	})
}
