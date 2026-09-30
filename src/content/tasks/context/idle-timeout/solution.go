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
	ctx, cancelCause := context.WithCancelCause(parent)
	// Таймер без горутины: функция выполнится в своей горутине по срабатыванию.
	timer := time.AfterFunc(idle, func() { cancelCause(ErrIdle) })
	// Контекст кончился по любой причине — таймер больше не нужен.
	stop := context.AfterFunc(ctx, func() { timer.Stop() })

	touch = func() {
		if ctx.Err() != nil {
			return // отменённый контекст не продлеваем
		}
		// Reset перезаводит таймер; если он успел сработать, контекст
		// уже отменён навсегда, и Reset ничего не изменит.
		timer.Reset(idle)
	}
	cancel = func() {
		cancelCause(context.Canceled)
		stop()
		timer.Stop()
	}
	return ctx, touch, cancel
}
