package main

import "context"

// Limiter ограничивает число одновременно выполняемых функций.
type Limiter struct {
	slots chan struct{} // буферизованный канал как семафор
}

// NewLimiter создаёт ограничитель на n одновременных вызовов (n >= 1).
func NewLimiter(n int) *Limiter {
	return &Limiter{slots: make(chan struct{}, n)}
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
	// select выбирает случайно среди готовых веток: без этой проверки
	// отменённый ctx при свободном слоте иногда пропустил бы fn.
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case l.slots <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	// Освобождение — сразу после захвата и только через defer:
	// паника fn иначе навсегда унесёт слот.
	defer func() { <-l.slots }()
	return fn(ctx)
}
