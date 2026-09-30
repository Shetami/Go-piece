package main

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("очередь закрыта")

// Job — задание. Получает контекст сервера: он отменяется только при
// принудительной остановке (Shutdown не успел дождаться очереди).
type Job func(ctx context.Context)

type Server struct {
	jobs chan Job
	quit chan struct{} // закрыт — Shutdown начался

	mu     sync.RWMutex // Submit держит RLock, пока может писать в jobs
	closed bool

	base       context.Context
	cancelBase context.CancelFunc
	wg         sync.WaitGroup
	done       chan struct{} // закрыт — все воркеры вышли
	once       sync.Once
}

// NewServer запускает workers воркеров и очередь на capacity заданий.
func NewServer(workers, capacity int) *Server {
	s := &Server{
		jobs: make(chan Job, capacity),
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
	s.base, s.cancelBase = context.WithCancel(context.Background())
	for range workers {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for job := range s.jobs { // дочитываем очередь до конца
				if s.base.Err() != nil {
					continue // принудительная остановка: оставшееся выбрасываем
				}
				job(s.base)
			}
		}()
	}
	go func() {
		s.wg.Wait()
		close(s.done)
	}()
	return s
}

// Submit кладёт задание в очередь. Очередь полна — ждёт места, пока жив
// ctx (иначе ctx.Err()). После начала Shutdown — ErrClosed, в том числе
// для Submit, который в этот момент ждал места. Никогда не паникует.
// Принятое (nil) задание при Shutdown без истечения его ctx будет выполнено.
func (s *Server) Submit(ctx context.Context, job Job) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ErrClosed
	}
	select {
	case <-s.quit: // Shutdown уже начался — не принимаем, даже если место есть
		return ErrClosed
	default:
	}
	select {
	case s.jobs <- job:
		return nil
	case <-s.quit:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Shutdown перестаёт принимать задания, даёт воркерам выполнить всё, что
// уже в очереди, и ждёт их. Если ctx кончился раньше — отменяет контекст
// заданий, выбрасывает невыполненные, дожидается выхода воркеров и
// возвращает ctx.Err(). Повторный вызов безопасен.
func (s *Server) Shutdown(ctx context.Context) error {
	s.once.Do(func() {
		close(s.quit) // будим Submit, ждущие места, — они отпустят RLock
		s.mu.Lock()   // ждём, пока все Submit выйдут из отправки в канал
		s.closed = true
		close(s.jobs) // теперь в jobs никто не пишет — закрывать безопасно
		s.mu.Unlock()
	})
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		s.cancelBase() // принудительно: текущие задания получают отмену
		<-s.done
		return ctx.Err()
	}
}
