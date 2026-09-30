package main

import (
	"sync"
	"time"
)

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
	mu      sync.Mutex
	rate    float64
	burst   float64
	now     func() time.Time
	buckets map[string]*tokenBucket
}

type tokenBucket struct {
	tokens float64   // дробное число токенов: доли копятся между вызовами
	last   time.Time // когда tokens были актуальны
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{rate: rate, burst: float64(burst), now: now, buckets: make(map[string]*tokenBucket)}
}

// refill доливает ведро до момента t.
func (l *Limiter) refill(b *tokenBucket, t time.Time) {
	if el := t.Sub(b.last).Seconds(); el > 0 {
		b.tokens = min(l.burst, b.tokens+el*l.rate)
	}
	b.last = t
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	t := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &tokenBucket{tokens: l.burst, last: t}
		l.buckets[key] = b
	}
	l.refill(b, t)
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (l *Limiter) Cleanup() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	t := l.now()
	n := 0
	for k, b := range l.buckets { // удалять во время range по мапе можно
		l.refill(b, t)
		if b.tokens >= l.burst {
			delete(l.buckets, k)
			n++
		}
	}
	return n
}

func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
