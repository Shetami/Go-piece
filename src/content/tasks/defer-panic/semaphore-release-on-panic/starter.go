package main

import "context"

// Limiter ограничивает число одновременно выполняемых функций.
type Limiter struct {
	// ваши поля
}

// NewLimiter создаёт ограничитель на n одновременных вызовов (n >= 1).
func NewLimiter(n int) *Limiter {
	// ваш код
	return &Limiter{}
}

// Do ждёт свободный слот и выполняет fn(ctx).
//   - Одновременно выполняется не больше n функций.
//   - Если ctx отменили, пока ждали слот, fn не вызывается, Do возвращает
//     ctx.Err(). Уже отменённый ctx — тоже не вызывать fn, даже если слот
//     свободен.
//   - Слот освобождается на любом пути: fn вернула ошибку, fn запаниковала
//     (паника летит дальше с тем же значением).
//   - Результат — ошибка fn.
func (l *Limiter) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// ваш код
	return fn(ctx)
}
