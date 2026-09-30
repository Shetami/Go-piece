package main

import "time"

// Backoff возвращает паузу перед повтором номер attempt (1, 2, 3, …)
// или Stop, если повторять больше не надо.
type Backoff func(attempt int) time.Duration

// Stop — «больше не повторять».
const Stop time.Duration = -1

// Exponential: base, 2·base, 4·base, … (attempt < 1 считается как 1).
// При переполнении — насыщение до math.MaxInt64, а не мусор.
func Exponential(base time.Duration) Backoff {
	// ваш код
	return func(attempt int) time.Duration { return 0 }
}

// Capped ограничивает паузы сверху значением limit. Stop проходит как есть.
func Capped(b Backoff, limit time.Duration) Backoff {
	// ваш код
	return b
}

// Jitter уменьшает паузу d на случайную долю: результат d − d·frac·r,
// где r = rnd() из [0, 1), то есть лежит в (d·(1−frac), d].
// frac вне [0, 1] прижимается к границам. Stop проходит как есть.
func Jitter(b Backoff, frac float64, rnd func() float64) Backoff {
	// ваш код
	return b
}

// MaxAttempts: попытки с номером больше n получают Stop.
func MaxAttempts(b Backoff, n int) Backoff {
	// ваш код
	return b
}
