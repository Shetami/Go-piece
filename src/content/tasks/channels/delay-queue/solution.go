package main

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type delayItem[T any] struct {
	v   T
	at  time.Time
	seq uint64 // порядок добавления — для равных at
}

type delayHeap[T any] []delayItem[T]

func (h delayHeap[T]) Len() int { return len(h) }
func (h delayHeap[T]) Less(i, j int) bool {
	if !h[i].at.Equal(h[j].at) {
		return h[i].at.Before(h[j].at)
	}
	return h[i].seq < h[j].seq
}
func (h delayHeap[T]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *delayHeap[T]) Push(x any)   { *h = append(*h, x.(delayItem[T])) }
func (h *delayHeap[T]) Pop() any {
	old := *h
	it := old[len(old)-1]
	*h = old[:len(old)-1]
	return it
}

// DelayQueue — очередь отложенных задач (повторы с задержкой, напоминания,
// истечение TTL). Элемент становится доступен через delay после Put.
// Безопасна для многих производителей и потребителей.
type DelayQueue[T any] struct {
	mu      sync.Mutex
	items   delayHeap[T]
	seq     uint64
	changed chan struct{} // закрывается при каждом Put: «пересчитайте ожидание»
}

func NewDelayQueue[T any]() *DelayQueue[T] {
	return &DelayQueue[T]{changed: make(chan struct{})}
}

// Put добавляет v, который станет доступен через delay (delay <= 0 — сразу).
func (q *DelayQueue[T]) Put(v T, delay time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.seq++
	heap.Push(&q.items, delayItem[T]{v: v, at: time.Now().Add(delay), seq: q.seq})
	// Будим всех ждущих: новый элемент может быть раньше того,
	// под который они завели таймер. Один сигнал в канал с буфером
	// разбудил бы только одного из нескольких потребителей.
	close(q.changed)
	q.changed = make(chan struct{})
}

// Pop ждёт и возвращает элемент с самым ранним временем готовности; при
// равном времени — тот, что добавлен раньше. Если, пока Pop ждёт, добавлен
// элемент, готовый раньше, Pop должен вернуть его вовремя. Если ctx
// отменён раньше — нулевое значение и ctx.Err().
func (q *DelayQueue[T]) Pop(ctx context.Context) (T, error) {
	for {
		q.mu.Lock()
		changed := q.changed // берём под тем же замком, что и смотрим на кучу
		var wait <-chan time.Time
		var timer *time.Timer
		if len(q.items) > 0 {
			d := time.Until(q.items[0].at)
			if d <= 0 {
				it := heap.Pop(&q.items).(delayItem[T])
				q.mu.Unlock()
				return it.v, nil
			}
			timer = time.NewTimer(d)
			wait = timer.C
		}
		q.mu.Unlock()

		select {
		case <-wait: // голова созрела — на следующем круге заберём
		case <-changed: // пришёл новый элемент — пересчитать
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			var zero T
			return zero, ctx.Err()
		}
		if timer != nil {
			timer.Stop()
		}
	}
}

// Len — сколько элементов в очереди (готовых и нет).
func (q *DelayQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}
