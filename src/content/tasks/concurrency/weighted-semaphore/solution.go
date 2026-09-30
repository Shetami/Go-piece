package main

import (
	"container/list"
	"context"
	"errors"
	"sync"
)

var ErrTooLarge = errors.New("semaphore: request exceeds size")

// Weighted — семафор с весами на size единиц (как golang.org/x/sync/semaphore).
//
//   - Acquire(ctx, n) занимает n единиц, блокируясь, пока их не хватит.
//     n > size — сразу ErrTooLarge.
//   - Ожидающие обслуживаются строго по очереди (FIFO): пока первый в
//     очереди ждёт, следующие его не обгоняют, даже если им хватило бы.
//   - Если ctx отменён раньше, чем единицы выданы, Acquire возвращает
//     ctx.Err() и ничего не занимает, а очередь двигается дальше.
//   - TryAcquire(n) — без ожидания; false, если единиц мало или есть очередь.
//   - Release(n) возвращает единицы и будит тех, кому теперь хватает.
//     Вернуть больше, чем занято, — паника.
type Weighted struct {
	size    int64
	mu      sync.Mutex
	cur     int64
	waiters list.List // *waiter в порядке прихода
}

type waiter struct {
	n     int64
	ready chan struct{} // закрывается, когда единицы выданы
}

func NewWeighted(size int64) *Weighted { return &Weighted{size: size} }

func (s *Weighted) Acquire(ctx context.Context, n int64) error {
	if n > s.size {
		return ErrTooLarge
	}
	s.mu.Lock()
	if s.size-s.cur >= n && s.waiters.Len() == 0 {
		s.cur += n
		s.mu.Unlock()
		return nil
	}
	w := &waiter{n: n, ready: make(chan struct{})}
	elem := s.waiters.PushBack(w)
	s.mu.Unlock()

	select {
	case <-w.ready:
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		select {
		case <-w.ready:
			// Единицы выдали одновременно с отменой: отменять поздно,
			// отдаём их как успех — иначе они потеряются.
			s.mu.Unlock()
			return nil
		default:
		}
		isFront := s.waiters.Front() == elem
		s.waiters.Remove(elem)
		// Если уходит голова очереди, за ней могут стоять те, кому уже хватает.
		if isFront {
			s.notify()
		}
		s.mu.Unlock()
		return ctx.Err()
	}
}

func (s *Weighted) TryAcquire(n int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.size-s.cur >= n && s.waiters.Len() == 0 {
		s.cur += n
		return true
	}
	return false
}

func (s *Weighted) Release(n int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cur -= n
	if s.cur < 0 {
		panic("semaphore: released more than held")
	}
	s.notify()
}

// notify выдаёт единицы ожидающим с головы очереди, пока хватает.
// Вызывается под s.mu.
func (s *Weighted) notify() {
	for {
		front := s.waiters.Front()
		if front == nil {
			return
		}
		w := front.Value.(*waiter)
		if s.size-s.cur < w.n {
			return // голове не хватает — остальные ждут за ней (FIFO)
		}
		s.cur += w.n
		s.waiters.Remove(front)
		close(w.ready)
	}
}
