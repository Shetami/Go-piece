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
	// ваш код
	l.mu.Lock()
	defer l.mu.Unlock()
	time.Sleep(time.Until(l.next))
	l.next = time.Now().Add(l.every)
	return nil
}
