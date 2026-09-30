package main

import (
	"context"
	"sync"
)

// Queue — очередь с блокирующим Pop на sync.Cond.
type Queue[T any] struct {
	mu    sync.Mutex
	cond  *sync.Cond
	items []T
}

func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Push добавляет элемент в хвост и будит одного ожидающего.
func (q *Queue[T]) Push(v T) {
	q.mu.Lock()
	q.items = append(q.items, v)
	q.mu.Unlock()
	q.cond.Signal()
}

// Len — число элементов в очереди.
func (q *Queue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Pop забирает элемент из головы; если очередь пуста — ждёт, пока он
// появится или пока не отменят ctx. Если ctx отменён (в том числе заранее),
// возвращает нулевое значение и context.Cause(ctx), элемент не забирает.
// Не оставляет после себя горутин.
func (q *Queue[T]) Pop(ctx context.Context) (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 {
		q.cond.Wait() // ваш код: сейчас отмена сюда не доходит
	}
	v := q.items[0]
	q.items = q.items[1:]
	return v, nil
}
