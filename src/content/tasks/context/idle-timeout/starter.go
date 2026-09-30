package main

import (
	"context"
	"errors"
	"time"
)

var ErrIdle = errors.New("нет активности дольше допустимого")

// WithIdleTimeout возвращает контекст, который отменяется с причиной
// ErrIdle, если touch не вызывали дольше idle подряд. Каждый touch
// сдвигает срок на idle от момента вызова. Это скользящий таймаут
// простоя, а не общий дедлайн: пока активность есть, контекст живёт
// сколько угодно.
//   - отмена parent — отмена с причиной родителя;
//   - cancel — причина context.Canceled, таймер освобождается;
//   - touch после отмены ничего не делает и контекст не «воскрешает»;
//   - touch и cancel безопасно вызывать из разных горутин.
func WithIdleTimeout(parent context.Context, idle time.Duration) (ctx context.Context, touch func(), cancel context.CancelFunc) {
	// ваш код
	ctx, cancel = context.WithTimeoutCause(parent, idle, ErrIdle)
	return ctx, func() {}, cancel
}
