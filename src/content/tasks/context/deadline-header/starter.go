package main

import (
	"context"
	"time"
)

const TimeoutHeader = "X-Timeout-Ms"

// Inject записывает в h[TimeoutHeader] оставшееся время ctx минус reserve
// (запас на сеть и сериализацию), в целых миллисекундах с округлением вниз.
// Отрицательное — "0". Нет дедлайна — заголовок удаляется (если был).
func Inject(ctx context.Context, h map[string]string, reserve time.Duration) {
	// ваш код
}

// Extract создаёт контекст запроса на стороне сервера:
//   - заголовок — неотрицательное целое число миллисекунд: таймаут
//     min(это значение, limit); "0" — контекст уже истёк;
//   - заголовка нет или он некорректен ("abc", "-5", "1.5", "",
//     число за пределами int64) — таймаут limit;
//   - дедлайн родителя, если он раньше, сохраняется.
//
// Огромные корректные значения не должны переполнять time.Duration.
func Extract(parent context.Context, h map[string]string, limit time.Duration) (context.Context, context.CancelFunc) {
	// ваш код
	return context.WithCancel(parent)
}
