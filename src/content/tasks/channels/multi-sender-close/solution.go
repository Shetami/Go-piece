package main

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("sink closed")

// Sink — канал, в который пишут многие горутины, а закрывает тот, кто
// решил, что хватит (например, при остановке сервиса).
type Sink[T any] struct {
	out  chan T
	done chan struct{} // сигнал «закрываемся» для ждущих Send

	// RLock держит каждый Send на время отправки, Lock — Close перед
	// close(out). Так out закрывается, только когда в него никто не пишет.
	mu     sync.RWMutex
	closed bool
	once   sync.Once
}

func NewSink[T any]() *Sink[T] {
	return &Sink[T]{out: make(chan T), done: make(chan struct{})}
}

// Out — канал для чтения. Каждый вызов — один и тот же канал. Он
// закрывается, когда Close вызван и все начатые Send вернулись.
func (s *Sink[T]) Out() <-chan T {
	return s.out
}

// Send отправляет v, дожидаясь читателя. Возвращает nil, если значение
// принято (и тогда читатель Out его получит), ErrClosed — если Sink
// закрыт до или во время ожидания, ctx.Err() — если отменён ctx.
// Send никогда не паникует.
func (s *Sink[T]) Send(ctx context.Context, v T) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ErrClosed // out уже закрыт — писать в него нельзя
	}
	select {
	case s.out <- v:
		return nil
	case <-s.done:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close закрывает Sink: ждущие Send возвращают ErrClosed, новые — тоже.
// Возвращается, когда Out закрыт. Повторный и конкурентный Close безопасны.
func (s *Sink[T]) Close() {
	s.once.Do(func() {
		// Сначала будим ждущих: они держат RLock, и без этого Lock
		// ниже ждал бы их вечно (если никто не читает Out).
		close(s.done)
		s.mu.Lock()
		s.closed = true
		close(s.out)
		s.mu.Unlock()
	})
}
