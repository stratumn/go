// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package batchfossilizer

import (
	"io/ioutil"
	"os"
	"testing"
)

func BenchmarkFossilizeMaxLeaves100(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100})
}

func BenchmarkFossilizeMaxLeaves1000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000})
}

func BenchmarkFossilizeMaxLeaves10000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 10000})
}

func BenchmarkFossilizeMaxLeaves100000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100000})
}

func BenchmarkFossilizeMaxLeaves1000000(b *testing.B) {
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000000})
}

func BenchmarkFossilizeMaxLeavesPath100(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100, Path: path})
}

func BenchmarkFossilizeMaxLeavesPath1000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000, Path: path})
}

func BenchmarkFossilizeMaxLeavesPath10000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 10000, Path: path})
}

func BenchmarkFossilizeMaxLeavesPath100000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100000, Path: path})
}

func BenchmarkFossilizeMaxLeavesPath1000000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000000, Path: path})
}

func BenchmarkFossilizeMaxLeavesPathFSync100(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100, Path: path, FSync: true})
}

func BenchmarkFossilizeMaxLeavesPathFSync1000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000, Path: path, FSync: true})
}

func BenchmarkFossilizeMaxLeavesPathFSync10000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 10000, Path: path, FSync: true})
}

func BenchmarkFossilizeMaxLeavesPathFSync100000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 100000, Path: path, FSync: true})
}

func BenchmarkFossilizeMaxLeavesPathFSync1000000(b *testing.B) {
	path, err := ioutil.TempDir("", "batchfossilizer")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)
	benchmarkFossilize(b, &Config{Interval: interval, MaxLeaves: 1000000, Path: path, FSync: true})
}
