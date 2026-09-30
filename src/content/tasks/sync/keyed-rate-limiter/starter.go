package main

import "time"

// Limiter — rate limiter «token bucket» с отдельным ведром на каждого
// клиента (по API-ключу или IP).
//
// Ведро вмещает burst токенов и пополняется со скоростью rate токенов в
// секунду, непрерывно (дробные токены копятся). Новое ведро — полное.
// Allow тратит один токен, если он есть. Если часы now() пошли назад,
// ведро не пополняется и не ломается.
type Limiter struct {
	// ваши поля
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	// ваш код
	return &Limiter{}
}

// Allow сообщает, можно ли пропустить запрос клиента key.
func (l *Limiter) Allow(key string) bool {
	// ваш код
	return true
}

// Cleanup удаляет вёдра, которые к текущему моменту пополнились до
// burst: они неотличимы от новых. Возвращает, сколько удалено.
func (l *Limiter) Cleanup() int {
	// ваш код
	return 0
}

// Len — сколько вёдер сейчас хранится.
func (l *Limiter) Len() int {
	// ваш код
	return 0
}
