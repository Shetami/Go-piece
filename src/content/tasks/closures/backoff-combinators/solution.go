package main

import (
	"math"
	"time"
)

// Backoff возвращает паузу перед повтором номер attempt (1, 2, 3, …)
// или Stop, если повторять больше не надо.
type Backoff func(attempt int) time.Duration

// Stop — «больше не повторять».
const Stop time.Duration = -1

// Exponential: base, 2·base, 4·base, … (attempt < 1 считается как 1).
// При переполнении — насыщение до math.MaxInt64, а не мусор.
func Exponential(base time.Duration) Backoff {
	return func(attempt int) time.Duration {
		d := base
		for i := 1; i < attempt; i++ {
			if d > math.MaxInt64/2 {
				return math.MaxInt64 // дальше удваивать некуда
			}
			d *= 2
		}
		return d
	}
}

// Capped ограничивает паузы сверху значением limit. Stop проходит как есть.
func Capped(b Backoff, limit time.Duration) Backoff {
	return func(attempt int) time.Duration {
		d := b(attempt)
		if d == Stop {
			return Stop
		}
		return min(d, limit)
	}
}

// Jitter уменьшает паузу d на случайную долю: результат d − d·frac·r,
// где r = rnd() из [0, 1), то есть лежит в (d·(1−frac), d].
// frac вне [0, 1] прижимается к границам. Stop проходит как есть.
func Jitter(b Backoff, frac float64, rnd func() float64) Backoff {
	frac = min(max(frac, 0), 1)
	return func(attempt int) time.Duration {
		d := b(attempt)
		if d <= 0 {
			return d // и Stop, и ноль джиттерить нечего
		}
		// Считаем только вычитаемое: float64(math.MaxInt64) == 2^63, и
		// обратное преобразование произведения во Duration переполнилось бы.
		x := float64(d) * frac * rnd()
		if x >= float64(d) {
			return 0
		}
		return d - time.Duration(x)
	}
}

// MaxAttempts: попытки с номером больше n получают Stop.
func MaxAttempts(b Backoff, n int) Backoff {
	return func(attempt int) time.Duration {
		if attempt > n {
			return Stop
		}
		return b(attempt)
	}
}
