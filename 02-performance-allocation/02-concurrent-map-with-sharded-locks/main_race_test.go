package main

import (
	"strconv"
	"sync"
	"testing"
)


func TestShardedMapRace_Int(t *testing.T) {
	m := NewShardedMap[int, int](32, hashInt)

	const (
		writers    = 4
		readers    = 4
		deleters   = 2
		iterations = 50_000
		keySpace   = 10_000
	)

	var wg sync.WaitGroup
	wg.Add(writers + readers + deleters)

	for w := 0; w < writers; w++ {
		w := w
		go func() {
			defer wg.Done()
			base := w * iterations
			for i := 0; i < iterations; i++ {
				k := (base + i) % keySpace
				m.Set(k, i)
			}
		}()
	}

	for r := 0; r < readers; r++ {
		r := r
		go func() {
			defer wg.Done()
			base := r * iterations
			for i := 0; i < iterations; i++ {
				k := (base + i) % keySpace
				_, _ = m.Get(k)
			}
		}()
	}

	for d := 0; d < deleters; d++ {
		d := d
		go func() {
			defer wg.Done()
			base := d * iterations
			for i := 0; i < iterations; i++ {
				k := (base + i) % keySpace
				m.Delete(k)
			}
		}()
	}

	wg.Wait()
}

func TestShardedMapRace_StringWithKeys(t *testing.T) {
	m := NewShardedMap[string, int](32, hashFNV)

	const (
		writers    = 4
		readers    = 4
		deleters   = 2
		keyReaders = 2
		iterations = 30_000
		keySpace   = 5_000
	)

	var wg sync.WaitGroup
	wg.Add(writers + readers + deleters + keyReaders)

	for w := 0; w < writers; w++ {
		w := w
		go func() {
			defer wg.Done()
			base := w * iterations
			for i := 0; i < iterations; i++ {
				k := strconv.Itoa((base + i) % keySpace)
				m.Set(k, i)
			}
		}()
	}

	for r := 0; r < readers; r++ {
		r := r
		go func() {
			defer wg.Done()
			base := r * iterations
			for i := 0; i < iterations; i++ {
				k := strconv.Itoa((base + i) % keySpace)
				_, _ = m.Get(k)
			}
		}()
	}

	for d := 0; d < deleters; d++ {
		d := d
		go func() {
			defer wg.Done()
			base := d * iterations
			for i := 0; i < iterations; i++ {
				k := strconv.Itoa((base + i) % keySpace)
				m.Delete(k)
			}
		}()
	}

	for kr := 0; kr < keyReaders; kr++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 2_000; i++ {
				_ = m.Keys()
			}
		}()
	}

	wg.Wait()
}