package main

import (
	"strconv"
	"sync"
	"testing"
)

func BenchmarkShardedMapSet_1Shard(b *testing.B) {
	benchmarkShardedMapSet(b, 1)
}

func BenchmarkShardedMapSet_50Shards(b *testing.B) {
	benchmarkShardedMapSet(b, 50)
}


func BenchmarkShardedMapSet_64Shards(b *testing.B) {
	benchmarkShardedMapSet(b, 64)
}


func BenchmarkShardedMapSet_128Shards(b *testing.B) {
	benchmarkShardedMapSet(b, 128)
}


func benchmarkShardedMapSet(b *testing.B, shardCount int) {
	const workers = 8

	m := NewShardedMap[string, int](shardCount, hashFNV)

	b.ReportAllocs()
	b.ResetTimer()

	var wg sync.WaitGroup
	wg.Add(workers)

	opsPerWorker := b.N / workers
	remainder := b.N % workers

	for w := 0; w < workers; w++ {
		w := w

		n := opsPerWorker
		if w < remainder {
			n++
		}

		go func() {
			defer wg.Done()

			start := w * opsPerWorker
			for i := 0; i < n; i++ {
				// последовательные уникальные ключи
				key := strconv.Itoa(start + i)
				m.Set(key, i)
			}
		}()
	}

	wg.Wait()
}