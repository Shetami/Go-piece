package main

import "time"

// Limiter — token bucket отдельно для каждого ключа (пользователя, IP, API-ключа).
// Потокобезопасен. Время берётся только из now.
//
//   - Ведро вмещает burst токенов и пополняется со скоростью rate токенов
//     в секунду — непрерывно, в том числе дробными долями, но не выше burst.
//   - Новый ключ начинает с полным ведром.
//   - Allow(key) забирает токен и возвращает true, если в ведре есть хотя бы
//     один целый токен; иначе false, и ведро не меняется (кроме пополнения).
//   - Cleanup удаляет ключи, чьи вёдра уже снова полны: их состояние не
//     отличить от нового ключа. Возвращает число удалённых.
//   - Len — сколько ключей хранится.
type Limiter struct {
	// ваши поля
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{}
}

func (l *Limiter) Allow(key string) bool {
	// ваш код
	return false
}

func (l *Limiter) Cleanup() int {
	// ваш код
	return 0
}

func (l *Limiter) Len() int {
	// ваш код
	return 0
}
