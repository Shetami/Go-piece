package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrDeadlineTooSoon = errors.New("лимитер: слот освободится только после дедлайна")

// Limiter пропускает не больше одного события в every. Первое — сразу.
type Limiter struct {
	every time.Duration
	mu    sync.Mutex
	next  time.Time // ближайший свободный слот
}

func NewLimiter(every time.Duration) *Limiter {
	return &Limiter{every: every}
}

// Wait резервирует ближайший свободный слот и ждёт его наступления.
//   - ctx уже отменён — context.Cause(ctx), слот не резервируется;
//   - у ctx есть дедлайн раньше слота — сразу ErrDeadlineTooSoon, без
//     ожидания и без резервирования;
//   - ctx отменили во время ожидания — context.Cause(ctx); слот
//     возвращается, если после него никто не успел зарезервировать следующий.
//
// Безопасен для вызова из многих горутин; каждый слот достаётся одному.
func (l *Limiter) Wait(ctx context.Context) error {
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}
	l.mu.Lock()
	now := time.Now()
	slot := l.next
	if slot.Before(now) {
		slot = now
	}
	// Заранее знаем, что не успеем, — не занимаем чужой слот и не спим зря.
	if dl, ok := ctx.Deadline(); ok && dl.Before(slot) {
		l.mu.Unlock()
		return ErrDeadlineTooSoon
	}
	l.next = slot.Add(l.every)
	l.mu.Unlock()

	wait := time.Until(slot)
	if wait <= 0 {
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		l.mu.Lock()
		if l.next.Equal(slot.Add(l.every)) { // мы последние в очереди — отдаём слот
			l.next = slot
		}
		l.mu.Unlock()
		return context.Cause(ctx)
	}
}
