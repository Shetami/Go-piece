package main

import (
	"container/heap"
	"sync"
	"time"
)

// FakeClock — часы для тестов: время идёт только в Advance.
// Безопасен для горутин.
type FakeClock struct {
	mu     sync.Mutex
	now    time.Time
	seq    int // порядковый номер планирования — для равных моментов
	timers timerHeap
}

type fakeTimer struct {
	at    time.Time
	seq   int
	f     func()
	index int  // позиция в куче, -1 — не в куче
	done  bool // сработал или отменён
}

func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{now: start}
}

// Now возвращает текущее время часов.
func (c *FakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// AfterFunc планирует f на момент Now()+d (d <= 0 — на текущий момент:
// f сработает при ближайшем Advance, даже Advance(0)).
// stop отменяет таймер и возвращает true, если f ещё не вызывалась и не
// была отменена; иначе false.
func (c *FakeClock) AfterFunc(d time.Duration, f func()) (stop func() bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &fakeTimer{at: c.now.Add(max(d, 0)), seq: c.seq, f: f}
	c.seq++
	heap.Push(&c.timers, t)
	// stop — замыкание над конкретным таймером: так его можно найти в
	// куче без идентификаторов и без сравнения функций.
	return func() bool {
		c.mu.Lock()
		defer c.mu.Unlock()
		if t.done {
			return false
		}
		t.done = true
		if t.index >= 0 {
			heap.Remove(&c.timers, t.index)
		}
		return true
	}
}

// Advance двигает время на d и вызывает созревшие f в порядке их времени,
// при равном времени — в порядке планирования. Пока выполняется f, Now()
// возвращает момент срабатывания её таймера. f может планировать и
// отменять таймеры; новые, созревшие до конца Advance, срабатывают в нём
// же. f вызывается не под блокировкой. После Advance Now() == старое + d.
func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	target := c.now.Add(max(d, 0))
	for len(c.timers) > 0 && !c.timers[0].at.After(target) {
		t := heap.Pop(&c.timers).(*fakeTimer)
		t.done = true
		c.now = t.at
		c.mu.Unlock()
		t.f() // без замка: f может звать Now, AfterFunc и stop
		c.mu.Lock()
	}
	c.now = target
	c.mu.Unlock()
}

// timerHeap — мин-куча по (at, seq).
type timerHeap []*fakeTimer

func (h timerHeap) Len() int { return len(h) }
func (h timerHeap) Less(i, j int) bool {
	if !h[i].at.Equal(h[j].at) {
		return h[i].at.Before(h[j].at)
	}
	return h[i].seq < h[j].seq
}
func (h timerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}
func (h *timerHeap) Push(x any) {
	t := x.(*fakeTimer)
	t.index = len(*h)
	*h = append(*h, t)
}
func (h *timerHeap) Pop() any {
	old := *h
	t := old[len(old)-1]
	old[len(old)-1] = nil
	*h = old[:len(old)-1]
	t.index = -1
	return t
}
