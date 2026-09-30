package main

import (
	"container/list"
	"context"
	"errors"
	"sync"
)

// ErrTooLarge — запрос больше ёмкости семафора, он не выполнится никогда.
var ErrTooLarge = errors.New("request exceeds semaphore size")

type waiter struct {
	n     int64
	ready chan struct{} // закрывается, когда единицы выданы
}

// Weighted — семафор с весами: задача занимает столько единиц, сколько
// ей нужно (например, мегабайт памяти под обработку файла).
//
// Порядок — FIFO: пока в очереди кто-то ждёт, новый запрос встаёт за ним,
// даже если для нового места хватает. Иначе большой запрос будет вечно
// голодать за потоком маленьких.
type Weighted struct {
	size    int64
	mu      sync.Mutex
	cur     int64
	waiters list.List // *waiter, по порядку прихода
}

func NewWeighted(size int64) *Weighted {
	return &Weighted{size: size}
}

// Acquire занимает n единиц, ожидая их освобождения. При отмене ctx
// возвращает ctx.Err() и ничего не занимает. n > size — сразу ErrTooLarge.
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
		defer s.mu.Unlock()
		select {
		case <-w.ready:
			// Единицы выдали одновременно с отменой — возвращаем их,
			// иначе они потеряны навсегда.
			s.cur -= n
			s.notifyLocked()
		default:
			front := s.waiters.Front() == elem
			s.waiters.Remove(elem)
			// Мы стояли первыми и загораживали очередь — теперь за нами
			// могут пройти те, кому места уже хватает.
			if front {
				s.notifyLocked()
			}
		}
		return ctx.Err()
	}
}

// TryAcquire занимает n единиц без ожидания, если это можно сделать сразу
// и в очереди никого нет.
func (s *Weighted) TryAcquire(n int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.size-s.cur >= n && s.waiters.Len() == 0 {
		s.cur += n
		return true
	}
	return false
}

// Release возвращает n единиц. Вернуть больше, чем занято, — паника.
func (s *Weighted) Release(n int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n > s.cur {
		panic("semaphore: released more than held")
	}
	s.cur -= n
	s.notifyLocked()
}

// notifyLocked выдаёт единицы ждущим строго по порядку, пока хватает места.
// Первый, кому не хватает, останавливает раздачу — даже если следующим
// хватило бы: в этом и состоит защита от голодания.
func (s *Weighted) notifyLocked() {
	for {
		e := s.waiters.Front()
		if e == nil {
			return
		}
		w := e.Value.(*waiter)
		if s.size-s.cur < w.n {
			return
		}
		s.cur += w.n
		s.waiters.Remove(e)
		close(w.ready)
	}
}
