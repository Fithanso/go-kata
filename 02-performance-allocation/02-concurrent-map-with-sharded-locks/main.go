package main

import (
	"fmt"
	"sync"
)

func New[V any]() ShardedMap[string, V] {
	return NewShardedMap[string, V](32, hashFNV)
} 


type ShardedMap[K comparable, V any] struct {
	shards   []Shard[K, V]
	hashFunc func(key K) uint64
}

func NewShardedMap[K comparable, V any] (count int, hashFunc func (key K) uint64) ShardedMap[K, V] {
	
	if count <= 0 {
		panic("count must be positive")
	}
	
	if hashFunc == nil {
		panic("hashFunc must not be nil")
	}
	
	shards := make([]Shard[K, V], count)
	
	for i := 0; i < count; i++ {
		shards[i] = Shard[K, V]{items: make(map[K]V)}
	}
	
	sm := ShardedMap[K, V]{
		shards,
		hashFunc,
	}
	return sm
}

type Shard[K comparable, V any]  struct {
	items map[K]V
	mu sync.RWMutex
}


func (s *ShardedMap[K, V]) Get(key K) (V, bool) {
	
	shardIndex := s.hashFunc(key) % uint64(len(s.shards))
	shard := &s.shards[shardIndex]
	
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	
	i, ok := shard.items[key]
	
	return i, ok
}

func(s *ShardedMap[K, V]) Set(key K, value V) {
	shardIndex := s.hashFunc(key) % uint64(len(s.shards))
	shard := &s.shards[shardIndex]
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.items[key] = value
}

func (s *ShardedMap[K, V]) Delete(key K) {
	shardIndex := s.hashFunc(key) % uint64(len(s.shards))
	shard := &s.shards[shardIndex]
	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.items, key)
}

func (s *ShardedMap[K, V]) Keys() []K {
	keys := make([]K, 0)
	for i := range(s.shards) {
		s := &s.shards[i]
		s.mu.RLock()
		for k := range(s.items) {
			keys = append(keys, k)
		}
		s.mu.RUnlock()
	}
	return keys
}



func hashFNV(key string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	var h uint64 = offset64
	for i := 0; i < len(key); i++ {
		h ^= uint64(key[i])
		h *= prime64
	}
	return h
}

func hashInt(key int) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	var h uint64 = offset64
	x := uint64(key)

	for i := 0; i < 8; i++ {
		h ^= x & 0xff
		h *= prime64
		x >>= 8
	}

	return h
}

func main() {
	fmt.Println(hashFNV("gavno"))
}