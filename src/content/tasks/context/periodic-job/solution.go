package main

import (
	"context"
	"fmt"
	"time"
)

// Every вызывает f сразу, затем раз в interval, пока ctx жив.
//   - вызовы f не перекрываются: следующий начинается только после
//     окончания предыдущего;
//   - если f работала дольше interval, пропущенные тики не копятся: после
//     долгого вызова — не больше одного вызова «вдогонку»;
//   - f получает ctx; ошибка f останавливает Every и возвращается
//     обёрнутой (errors.Is работает);
//   - отмена ctx — Every сразу возвращает context.Cause(ctx), в том числе
//     посреди ожидания следующего тика.
func Every(ctx context.Context, interval time.Duration, f func(context.Context) error) error {
	// Тикер сам сбрасывает лишние тики: в его канале буфер на один.
	t := time.NewTicker(interval)
	defer t.Stop()
	for n := 1; ; n++ {
		if ctx.Err() != nil { // select ниже мог выбрать тик, хотя ctx уже отменён
			return context.Cause(ctx)
		}
		if err := f(ctx); err != nil {
			return fmt.Errorf("every: запуск %d: %w", n, err)
		}
		select {
		case <-t.C:
		case <-ctx.Done():
			return context.Cause(ctx)
		}
	}
}
