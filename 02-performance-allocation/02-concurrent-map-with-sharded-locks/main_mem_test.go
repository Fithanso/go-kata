package main

import (
	"fmt"
	"runtime"
	"testing"
)

const n = 1_000_000

func measureHeap(f func()) uint64 {
	runtime.GC()
	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	f()

	runtime.GC()
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	return after.HeapAlloc - before.HeapAlloc
}

func TestMemoryUsage(t *testing.T) {
	baseline := measureHeap(func() {
		m := make(map[int]interface{}, n)
		for i := 0; i < n; i++ {
			m[i] = i
		}
		runtime.KeepAlive(m)
	})

	sharded := measureHeap(func() {
		m := NewShardedMap[int, interface{}](32, hashInt)
		for i := 0; i < n; i++ {
			m.Set(i, i)
		}
		runtime.KeepAlive(m)
	})

	diff := int64(sharded) - int64(baseline)

	fmt.Printf("baseline map: %d bytes\n", baseline)
	fmt.Printf("sharded map:  %d bytes\n", sharded)
	fmt.Printf("difference:   %d bytes\n", diff)

	const limit = 50 * 1024 * 1024 // 50 MB
	if diff > limit {
		t.Fatalf("memory overhead too high: got %d bytes, want <= %d", diff, limit)
	}
}