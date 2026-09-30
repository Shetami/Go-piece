package main

import (
	"sync"
	"time"
)

// NewLimiter — токен-бакет в замыкании.
// Вместимость — burst токенов, бакет пополняется со скоростью rate токенов
// в секунду (дробные доли копятся), но не выше burst. Изначально бакет полон.
// now — источник времени; если он вернул момент раньше предыдущего,
// считайте, что время не шло.
// allow забирает один токен и возвращает true, либо возвращает false,
// если целого токена нет. allow безопасна для вызова из многих горутин.
func NewLimiter(rate float64, burst int, now func() time.Time) (allow func() bool) {
	var (
		mu     sync.Mutex
		tokens = float64(burst)
		last   = now()
	)
	return func() bool {
		mu.Lock()
		defer mu.Unlock()
		t := now()
		// Пополняем дробно: целочисленное округление вместе со сдвигом last
		// теряло бы доли токена на частых вызовах.
		if elapsed := t.Sub(last); elapsed > 0 {
			tokens = min(float64(burst), tokens+elapsed.Seconds()*rate)
		}
		// Назад время не двигаем: иначе следующий вызов насчитает лишнее.
		if t.After(last) {
			last = t
		}
		if tokens < 1 {
			return false
		}
		tokens--
		return true
	}
}
