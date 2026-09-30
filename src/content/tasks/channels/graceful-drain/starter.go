package main

import (
	"context"
	"errors"
)

var (
	ErrOverloaded   = errors.New("overloaded")
	ErrShuttingDown = errors.New("shutting down")
)

// Server обрабатывает задачи: очередь на queueSize задач и workers
// воркеров, каждый вызывает handle.
type Server struct {
	// ваши поля
}

// NewServer запускает воркеров.
func NewServer(workers, queueSize int, handle func(job int)) *Server {
	// ваш код
	return &Server{}
}

// Submit ставит задачу в очередь, никогда не блокируясь: ErrOverloaded,
// если очередь полна; ErrShuttingDown, если Shutdown уже начат. nil
// означает, что задача точно будет обработана. Безопасен при любой
// конкуренции с Shutdown — никаких паник.
func (s *Server) Submit(job int) error {
	// ваш код
	return nil
}

// Shutdown перестаёт принимать задачи и ждёт, пока воркеры обработают
// всё, что уже в очереди, и завершатся. Если ctx кончился раньше —
// возвращает ctx.Err(), а воркеры доделывают в фоне. Повторный вызов
// безопасен и тоже ждёт завершения.
func (s *Server) Shutdown(ctx context.Context) error {
	// ваш код
	return nil
}
