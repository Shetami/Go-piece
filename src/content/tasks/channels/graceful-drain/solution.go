package main

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrOverloaded   = errors.New("overloaded")
	ErrShuttingDown = errors.New("shutting down")
)

// Server обрабатывает задачи: очередь на queueSize задач и workers
// воркеров, каждый вызывает handle.
type Server struct {
	queue chan int
	done  chan struct{} // закрывается, когда все воркеры вышли

	// Submit отправляет в queue под RLock, Shutdown закрывает её под Lock:
	// закрытие не может случиться посреди отправки.
	mu     sync.RWMutex
	closed bool
}

// NewServer запускает воркеров.
func NewServer(workers, queueSize int, handle func(job int)) *Server {
	s := &Server{
		queue: make(chan int, queueSize),
		done:  make(chan struct{}),
	}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// range дочитывает буфер и после close — очередь не теряется.
			for job := range s.queue {
				handle(job)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(s.done)
	}()
	return s
}

// Submit ставит задачу в очередь, никогда не блокируясь: ErrOverloaded,
// если очередь полна; ErrShuttingDown, если Shutdown уже начат. nil
// означает, что задача точно будет обработана. Безопасен при любой
// конкуренции с Shutdown — никаких паник.
func (s *Server) Submit(job int) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ErrShuttingDown
	}
	select {
	case s.queue <- job: // не блокируется, поэтому RLock держится недолго
		return nil
	default:
		return ErrOverloaded
	}
}

// Shutdown перестаёт принимать задачи и ждёт, пока воркеры обработают
// всё, что уже в очереди, и завершатся. Если ctx кончился раньше —
// возвращает ctx.Err(), а воркеры доделывают в фоне. Повторный вызов
// безопасен и тоже ждёт завершения.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.queue)
	}
	s.mu.Unlock()

	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
