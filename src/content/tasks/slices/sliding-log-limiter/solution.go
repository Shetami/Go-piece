package main

import (
	"sort"
	"time"
)

// Limiter пропускает не больше limit событий за любое скользящее окно длины window.
type Limiter struct {
	limit  int
	window time.Duration
	times  []time.Time // моменты пропущенных событий по возрастанию
}

// NewLimiter создаёт лимитер. limit <= 0 или window <= 0 — паника.
func NewLimiter(limit int, window time.Duration) *Limiter {
	if limit <= 0 || window <= 0 {
		panic("NewLimiter: limit и window должны быть положительными")
	}
	// Вместимость сразу limit: больше записей не бывает, append не перевыделит.
	return &Limiter{limit: limit, window: window, times: make([]time.Time, 0, limit)}
}

// Allow решает, пропустить ли событие в момент now; now от вызова к вызову
// не убывает. В окно попадают события из (now-window, now]: событие ровно
// window назад уже не считается. Отказанные события не записываются.
// Устаревшие записи удаляются, len(times) никогда не превышает limit,
// а в установившемся режиме Allow не выделяет память.
func (l *Limiter) Allow(now time.Time) bool {
	cut := now.Add(-l.window)
	// times отсортированы — первую живую запись ищем двоичным поиском.
	i := sort.Search(len(l.times), func(k int) bool { return l.times[k].After(cut) })
	if i > 0 {
		// Сдвигаем живые записи в начало того же массива. l.times = l.times[i:]
		// теряло бы вместимость, и append начал бы перевыделять память.
		n := copy(l.times, l.times[i:])
		l.times = l.times[:n]
	}
	if len(l.times) >= l.limit {
		return false
	}
	l.times = append(l.times, now)
	return true
}
