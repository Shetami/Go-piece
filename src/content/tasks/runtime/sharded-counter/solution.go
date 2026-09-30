package main

import (
	"math/rand/v2"
	"runtime"
	"sync/atomic"
)

// cacheLine — размер строки кэша процессора (x86-64, большинство ARM64).
const cacheLine = 64

// shard — счётчик одного шарда. Шарды лежат в слайсе подряд, и два
// соседних шарда не должны попадать в одну строку кэша: иначе ядра,
// обновляющие разные шарды, всё равно мешают друг другу (false sharing).
// Требование: unsafe.Sizeof(shard{}) кратен cacheLine.
type shard struct {
	n atomic.Int64
	_ [cacheLine - 8]byte // добивка до полной строки кэша
}

// Counter — счётчик для горячего пути (число запросов, байт), который
// увеличивают из тысяч горутин. Число шардов — runtime.GOMAXPROCS(0) на
// момент создания. Add не выделяет память. Безопасен для конкурентного
// использования.
type Counter struct {
	shards []shard
}

func NewCounter() *Counter {
	return &Counter{shards: make([]shard, runtime.GOMAXPROCS(0))}
}

func (c *Counter) Add(delta int64) {
	// Номер P горутине недоступен, поэтому шард выбираем случайно.
	// Глобальный генератор math/rand/v2 не берёт общий мьютекс и не аллоцирует.
	c.shards[rand.IntN(len(c.shards))].n.Add(delta)
}

// Value — сумма по всем шардам.
func (c *Counter) Value() int64 {
	var sum int64
	for i := range c.shards {
		sum += c.shards[i].n.Load()
	}
	return sum
}

// Shards — число шардов.
func (c *Counter) Shards() int { return len(c.shards) }
