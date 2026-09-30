package main

import (
	"errors"
	"time"
)

// ErrOpen — выключатель разомкнут, вызов не выполнялся.
var ErrOpen = errors.New("circuit open")

// Breaker — circuit breaker перед нестабильной зависимостью.
//
//   - closed: вызовы проходят; после threshold ошибок подряд (успех
//     обнуляет счёт) выключатель размыкается.
//   - open: вызовы сразу получают ErrOpen, fn не вызывается. Через
//     cooldown после размыкания следующий вызов становится пробным.
//   - half-open: идёт ровно один пробный вызов; все остальные в это время
//     получают ErrOpen. Успех пробы замыкает выключатель, ошибка —
//     снова размыкает на cooldown, отсчитанный от момента ошибки.
//
// fn выполняется без удержания внутренней блокировки: вызовы в
// состоянии closed идут параллельно.
type Breaker struct {
	// ваши поля
}

func NewBreaker(threshold int, cooldown time.Duration, now func() time.Time) *Breaker {
	// ваш код
	return &Breaker{}
}

// Call выполняет fn через выключатель и возвращает её ошибку или ErrOpen.
func (b *Breaker) Call(fn func() error) error {
	// ваш код
	return fn()
}

// State возвращает "closed", "open" или "half-open" (пока идёт проба).
func (b *Breaker) State() string {
	// ваш код
	return "closed"
}
