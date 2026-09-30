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
	// ваш код
	return nil
}
