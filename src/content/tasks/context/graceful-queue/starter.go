package main

import (
	"context"
	"errors"
)

var ErrClosed = errors.New("очередь закрыта")

// Job — задание. Получает контекст сервера: он отменяется только при
// принудительной остановке (Shutdown не успел дождаться очереди).
type Job func(ctx context.Context)

type Server struct {
	jobs chan Job
	// ваши поля
}

// NewServer запускает workers воркеров и очередь на capacity заданий.
func NewServer(workers, capacity int) *Server {
	s := &Server{jobs: make(chan Job, capacity)}
	for range workers {
		go func() {
			for job := range s.jobs {
				job(context.Background())
			}
		}()
	}
	return s
}

// Submit кладёт задание в очередь. Очередь полна — ждёт места, пока жив
// ctx (иначе ctx.Err()). После начала Shutdown — ErrClosed, в том числе
// для Submit, который в этот момент ждал места. Никогда не паникует.
// Принятое (nil) задание при Shutdown без истечения его ctx будет выполнено.
func (s *Server) Submit(ctx context.Context, job Job) error {
	// ваш код
	s.jobs <- job
	return nil
}

// Shutdown перестаёт принимать задания, даёт воркерам выполнить всё, что
// уже в очереди, и ждёт их. Если ctx кончился раньше — отменяет контекст
// заданий, выбрасывает невыполненные, дожидается выхода воркеров и
// возвращает ctx.Err(). Повторный вызов безопасен.
func (s *Server) Shutdown(ctx context.Context) error {
	// ваш код
	return nil
}
