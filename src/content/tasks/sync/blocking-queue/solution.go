package main

import (
	"errors"
	"sync"
)

// ErrClosed — очередь закрыта.
var ErrClosed = errors.New("queue closed")

// Queue — ограниченная блокирующая FIFO-очередь между производителями и
// потребителями (как буферизованный канал, но с Len и честным закрытием).
type Queue[T any] struct {
	mu       sync.Mutex
	notEmpty *sync.Cond // ждут потребители
	notFull  *sync.Cond // ждут производители
	buf      []T        // кольцевой буфер
	head, n  int
	closed   bool
}

// NewQueue создаёт очередь на capacity элементов (capacity >= 1).
func NewQueue[T any](capacity int) *Queue[T] {
	q := &Queue[T]{buf: make([]T, capacity)}
	// Два условия на одном мьютексе: Signal будит того, кого нужно.
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

// Put кладёт элемент, ожидая свободного места. Если очередь закрыта —
// до вызова или пока Put ждал места, — возвращает ErrClosed.
func (q *Queue[T]) Put(v T) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.n == len(q.buf) && !q.closed {
		q.notFull.Wait()
	}
	if q.closed {
		return ErrClosed
	}
	q.buf[(q.head+q.n)%len(q.buf)] = v
	q.n++
	q.notEmpty.Signal()
	return nil
}

// Take забирает самый старый элемент, ожидая, пока он появится.
// После Close отдаёт оставшиеся элементы, а когда их нет — (zero, false).
func (q *Queue[T]) Take() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.n == 0 && !q.closed {
		q.notEmpty.Wait()
	}
	var zero T
	if q.n == 0 { // закрыта и пуста
		return zero, false
	}
	v := q.buf[q.head]
	q.buf[q.head] = zero // не держим ссылку на отданный элемент
	q.head = (q.head + 1) % len(q.buf)
	q.n--
	q.notFull.Signal()
	return v, true
}

// Close закрывает очередь и будит всех ждущих. Повторный Close ничего не делает.
func (q *Queue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	// Broadcast: проснуться и увидеть закрытие должны все ждущие, а не один.
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
}

// Len — сколько элементов сейчас в очереди.
func (q *Queue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.n
}
