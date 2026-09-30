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
	// Cond.Wait не умеет слушать канал, поэтому отмена будит всех через
	// Broadcast. Под мьютексом — иначе Broadcast может проскочить между
	// нашей проверкой ctx.Err() и входом в Wait, и мы уснём навсегда.
	stop := context.AfterFunc(ctx, func() {
		q.mu.Lock()
		defer q.mu.Unlock()
		q.cond.Broadcast()
	})
	defer stop() // снимаем регистрацию, если ушли с элементом

	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		if ctx.Err() != nil {
			var zero T
			return zero, context.Cause(ctx)
		}
		if len(q.items) > 0 {
			v := q.items[0]
			var zero T
			q.items[0] = zero // не держим ссылку в хвосте массива
			q.items = q.items[1:]
			return v, nil
		}
		q.cond.Wait()
	}
}
