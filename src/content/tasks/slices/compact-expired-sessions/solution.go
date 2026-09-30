package main

import "time"

// Session — пользовательская сессия; Expires — момент, когда она истекает.
type Session struct {
	ID      string
	Expires time.Time
}

// DropExpired убирает из sessions истёкшие сессии (Expires <= now) и nil-указатели.
// Работает на месте: результат использует массив sessions, новых аллокаций нет,
// порядок оставшихся сохраняется. Ячейки массива за длиной результата
// обнуляются, чтобы сборщик мусора мог освободить выброшенные сессии.
func DropExpired(sessions []*Session, now time.Time) []*Session {
	// kept пишет в тот же массив, что читает range, но никогда не обгоняет его.
	kept := sessions[:0]
	for _, s := range sessions {
		if s != nil && s.Expires.After(now) {
			kept = append(kept, s)
		}
	}
	// Без этого хвост массива держит указатели на выброшенные сессии.
	clear(sessions[len(kept):])
	return kept
}
