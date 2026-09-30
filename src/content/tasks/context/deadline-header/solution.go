package main

import (
	"context"
	"math"
	"strconv"
	"time"
)

const TimeoutHeader = "X-Timeout-Ms"

// Inject записывает в h[TimeoutHeader] оставшееся время ctx минус reserve
// (запас на сеть и сериализацию), в целых миллисекундах с округлением вниз.
// Отрицательное — "0". Нет дедлайна — заголовок удаляется (если был).
func Inject(ctx context.Context, h map[string]string, reserve time.Duration) {
	dl, ok := ctx.Deadline()
	if !ok {
		delete(h, TimeoutHeader) // не тащим чужой устаревший заголовок дальше
		return
	}
	ms := max((time.Until(dl) - reserve).Milliseconds(), 0) // Milliseconds() усекает к нулю
	h[TimeoutHeader] = strconv.FormatInt(ms, 10)
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
	timeout := limit
	if s, ok := h[TimeoutHeader]; ok {
		if ms, err := strconv.ParseInt(s, 10, 64); err == nil && ms >= 0 {
			// ms * 1e6 может не влезть в int64 — сравниваем до умножения.
			if ms < math.MaxInt64/int64(time.Millisecond) {
				timeout = min(time.Duration(ms)*time.Millisecond, limit)
			}
		}
	}
	return context.WithTimeout(parent, timeout)
}
