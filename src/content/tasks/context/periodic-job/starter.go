package main

import (
	"context"
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
	// ваш код
	for range time.Tick(interval) {
		go f(ctx)
	}
	return nil
}
