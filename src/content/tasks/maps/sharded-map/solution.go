package main

import (
	"hash/maphash"
	"sync"
)

// ShardedMap — потокобезопасная мапа string → V, разбитая на n шардов,
// у каждого свой sync.RWMutex: операции над разными шардами не мешают друг
// другу. Шард выбирается по хэшу ключа. n < 1 считается как 1.
//
//   - Get, Set, Delete, Len — обычные.
//   - Update атомарно меняет значение ключа: fn получает текущее значение
//     (ok=false, если ключа нет) и возвращает новое; keep=false удаляет ключ.
//     Между чтением и записью никто не должен вклиниться. fn не обращается
//     к этой же мапе.
//   - Range вызывает fn для пар, пока fn возвращает true. Внутри fn можно
//     вызывать любые методы этой мапы (Set, Delete, Update…) — без дедлока.
//     Порядок обхода не важен.
type ShardedMap[V any] struct {
	seed   maphash.Seed
	shards []*shard[V]
}

type shard[V any] struct {
	mu sync.RWMutex
	m  map[string]V
}

func NewShardedMap[V any](n int) *ShardedMap[V] {
	n = max(n, 1)
	s := &ShardedMap[V]{seed: maphash.MakeSeed(), shards: make([]*shard[V], n)}
	for i := range s.shards {
		s.shards[i] = &shard[V]{m: make(map[string]V)}
	}
	return s
}

func (s *ShardedMap[V]) shardFor(key string) *shard[V] {
	return s.shards[maphash.String(s.seed, key)%uint64(len(s.shards))]
}

func (s *ShardedMap[V]) Get(key string) (V, bool) {
	sh := s.shardFor(key)
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	v, ok := sh.m[key]
	return v, ok
}

func (s *ShardedMap[V]) Set(key string, v V) {
	sh := s.shardFor(key)
	sh.mu.Lock()
	sh.m[key] = v
	sh.mu.Unlock()
}

func (s *ShardedMap[V]) Delete(key string) {
	sh := s.shardFor(key)
	sh.mu.Lock()
	delete(sh.m, key)
	sh.mu.Unlock()
}

func (s *ShardedMap[V]) Update(key string, fn func(v V, ok bool) (V, bool)) {
	sh := s.shardFor(key)
	// Чтение, fn и запись — под одной блокировкой, иначе Get+Set
	// теряет обновления, сделанные между ними другими горутинами.
	sh.mu.Lock()
	defer sh.mu.Unlock()
	old, ok := sh.m[key]
	if nv, keep := fn(old, ok); keep {
		sh.m[key] = nv
	} else {
		delete(sh.m, key)
	}
}

func (s *ShardedMap[V]) Len() int {
	n := 0
	for _, sh := range s.shards {
		sh.mu.RLock()
		n += len(sh.m)
		sh.mu.RUnlock()
	}
	return n
}

func (s *ShardedMap[V]) Range(fn func(key string, v V) bool) {
	type kv struct {
		k string
		v V
	}
	for _, sh := range s.shards {
		// Копируем шард под блокировкой, а fn зовём уже без неё:
		// иначе Set из fn возьмёт Lock того же шарда, пока мы держим RLock, —
		// дедлок (RWMutex не реентерабелен).
		sh.mu.RLock()
		batch := make([]kv, 0, len(sh.m))
		for k, v := range sh.m {
			batch = append(batch, kv{k, v})
		}
		sh.mu.RUnlock()
		for _, p := range batch {
			if !fn(p.k, p.v) {
				return
			}
		}
	}
}
