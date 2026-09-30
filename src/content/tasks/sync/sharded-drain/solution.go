package main

import (
	"hash/maphash"
	"sync"
)

const shardCount = 16

type shard struct {
	mu sync.Mutex
	m  map[string]int64
}

// Counter — счётчики по ключам для метрик: сотни горутин вызывают Add,
// а фоновый экспортёр раз в несколько секунд забирает накопленное через
// Drain. Чтобы Add по разным ключам не упирались в один мьютекс,
// раскидайте ключи по нескольким шардам.
type Counter struct {
	seed   maphash.Seed
	shards [shardCount]shard
}

func NewCounter() *Counter {
	c := &Counter{seed: maphash.MakeSeed()}
	for i := range c.shards {
		c.shards[i].m = make(map[string]int64)
	}
	return c
}

func (c *Counter) shardFor(key string) *shard {
	return &c.shards[maphash.String(c.seed, key)%shardCount]
}

// Add прибавляет delta к счётчику key.
func (c *Counter) Add(key string, delta int64) {
	s := c.shardFor(key)
	s.mu.Lock()
	s.m[key] += delta
	s.mu.Unlock()
}

// Get возвращает текущее значение счётчика (0, если ключа нет).
func (c *Counter) Get(key string) int64 {
	s := c.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[key]
}

// Drain возвращает всё накопленное с прошлого Drain и обнуляет счётчики.
// Прибавление, сделанное одновременно с Drain, не теряется и не
// учитывается дважды: оно попадает либо в этот Drain, либо в следующий.
// Результат — новая мапа, не nil (пустая, если ничего не накоплено).
func (c *Counter) Drain() map[string]int64 {
	out := make(map[string]int64)
	for i := range c.shards {
		s := &c.shards[i]
		// Забрать и обнулить шард — под одной блокировкой. Если сначала
		// скопировать все шарды, а потом отдельным проходом очистить,
		// всё, что пришло между проходами, пропадёт.
		s.mu.Lock()
		old := s.m
		s.m = make(map[string]int64, len(old))
		s.mu.Unlock()
		// Старую мапу больше никто не трогает — копируем без блокировки.
		for k, v := range old {
			out[k] = v
		}
	}
	return out
}
