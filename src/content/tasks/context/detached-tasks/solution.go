package main

import (
	"context"
	"sync"
	"time"
)

// Background запускает задачи, которые должны пережить запрос:
// запись аудита, отправку события, прогрев кэша.
type Background struct {
	timeout time.Duration
	limit   int

	mu      sync.Mutex
	closed  bool
	running int
	wg      sync.WaitGroup
}

func NewBackground(timeout time.Duration, limit int) *Background {
	return &Background{timeout: timeout, limit: limit}
}

// Go запускает f в отдельной горутине. Контекст f:
//   - НЕ отменяется вместе с ctx запроса, но видит его значения;
//   - имеет свой дедлайн: timeout от момента вызова Go.
//
// Одновременно работает не больше limit задач. Если все места заняты,
// Go не ждёт, а возвращает false (задача отброшена). После Shutdown —
// тоже false. true — задача запущена.
func (b *Background) Go(ctx context.Context, f func(context.Context)) bool {
	b.mu.Lock()
	if b.closed || b.running >= b.limit {
		b.mu.Unlock()
		return false
	}
	b.running++
	b.wg.Add(1) // под тем же мьютексом, что и closed: Shutdown не пропустит задачу
	b.mu.Unlock()

	// Значения — от запроса, отмена — нет, срок — свой.
	tctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), b.timeout)
	go func() {
		defer func() {
			cancel()
			b.mu.Lock()
			b.running--
			b.mu.Unlock()
			b.wg.Done()
		}()
		f(tctx)
	}()
	return true
}

// Shutdown запрещает новые задачи и ждёт текущие, но не дольше ctx.
// Если ctx кончился раньше — возвращает ctx.Err(). Повторный вызов безопасен.
func (b *Background) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
