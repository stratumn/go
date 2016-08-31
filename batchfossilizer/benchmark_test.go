// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"io/ioutil"
	"os"
	"testing"
)

func BenchmarkFossilize_MaxLeaves100(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 100})
}

func BenchmarkFossilize_MaxLeaves1000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 1000})
}

func BenchmarkFossilize_MaxLeaves10000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 10000})
}

func BenchmarkFossilize_MaxLeaves100000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 100000})
}

func BenchmarkFossilize_MaxLeaves1000000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 1000000})
}

func BenchmarkFossilize_Path_MaxLeaves100(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 100, Path: path})
}

func BenchmarkFossilize_Path_MaxLeaves1000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 1000, Path: path})
}

func BenchmarkFossilize_Path_MaxLeaves10000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 10000, Path: path})
}

func BenchmarkFossilize_Path_MaxLeaves100000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 100000, Path: path})
}

func BenchmarkFossilize_Path_MaxLeaves1000000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 1000000, Path: path})
}

func BenchmarkFossilize_FSync_MaxLeaves100(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 100, Path: path, FSync: true})
}

func BenchmarkFossilize_FSync_MaxLeaves1000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 1000, Path: path, FSync: true})
}

func BenchmarkFossilize_FSync_MaxLeaves10000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 10000, Path: path, FSync: true})
}

func BenchmarkFossilize_FSync_MaxLeaves100000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 100000, Path: path, FSync: true})
}

func BenchmarkFossilize_FSync_MaxLeaves1000000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatalf("ioutil.TempDir(): err: %s", err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, StopBatch: true, MaxLeaves: 1000000, Path: path, FSync: true})
}
