package main

import "time"

// Limiter пропускает не больше limit событий за любое скользящее окно длины window.
type Limiter struct {
	limit  int
	window time.Duration
	times  []time.Time // моменты пропущенных событий по возрастанию
}

// NewLimiter создаёт лимитер. limit <= 0 или window <= 0 — паника.
func NewLimiter(limit int, window time.Duration) *Limiter {
	// ваш код
	return &Limiter{}
}

// Allow решает, пропустить ли событие в момент now; now от вызова к вызову
// не убывает. В окно попадают события из (now-window, now]: событие ровно
// window назад уже не считается. Отказанные события не записываются.
// Устаревшие записи удаляются, len(times) никогда не превышает limit,
// а в установившемся режиме Allow не выделяет память.
func (l *Limiter) Allow(now time.Time) bool {
	// ваш код
	return false
}
