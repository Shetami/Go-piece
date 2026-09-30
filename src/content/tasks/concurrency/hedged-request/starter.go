package main

import (
	"context"
	"time"
)

// Hedged выполняет страхующие (hedged) запросы.
//
// Сразу вызывает call(ctx, 0). Если за delay ответа нет — запускает
// call(ctx, 1), ещё через delay — call(ctx, 2), и так до attempts попыток
// (attempts >= 1). Если попытка завершилась ошибкой, следующая запускается
// сразу, не дожидаясь delay.
//
// Возвращает первый успешный результат; контекст остальных попыток при этом
// отменяется. Если упали все — errors.Join ошибок в порядке номеров попыток
// (не в порядке их завершения). Если отменён ctx — сразу ctx.Err().
// Hedged не ждёт отменённые попытки, но и не оставляет горутин, навсегда
// заблокированных на отправке результата.
func Hedged[T any](ctx context.Context, attempts int, delay time.Duration,
	call func(ctx context.Context, i int) (T, error)) (T, error) {
	// ваш код
	var zero T
	return zero, nil
}
