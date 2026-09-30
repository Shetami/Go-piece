package main

import (
	"sync"
	"time"
)

type bucket struct {
	tokens float64   // дробные токены — иначе медленное пополнение теряется
	last   time.Time // когда tokens были актуальны
}

// Limiter — rate limiter «token bucket» с отдельным ведром на каждого
// клиента (по API-ключу или IP).
//
// Ведро вмещает burst токенов и пополняется со скоростью rate токенов в
// секунду, непрерывно (дробные токены копятся). Новое ведро — полное.
// Allow тратит один токен, если он есть. Если часы now() пошли назад,
// ведро не пополняется и не ломается.
type Limiter struct {
	rate  float64
	burst float64
	now   func() time.Time

	mu      sync.Mutex
	buckets map[string]*bucket
}

func NewLimiter(rate float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{rate: rate, burst: float64(burst), now: now, buckets: make(map[string]*bucket)}
}

// refill доливает токены за прошедшее время. Вызывается под l.mu.
func (l *Limiter) refill(b *bucket, now time.Time) {
	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens = min(l.burst, b.tokens+elapsed.Seconds()*l.rate)
		b.last = now
	}
	// elapsed <= 0 (часы назад): ничего не доливаем и не сдвигаем last
	// назад — иначе следующий вызов насчитал бы лишнее время.
}

// Allow сообщает, можно ли пропустить запрос клиента key.
func (l *Limiter) Allow(key string) bool {
	now := l.now()
	// Одна блокировка на всё «прочитать — пополнить — списать»: иначе две
	// горутины одного клиента потратят один и тот же токен.
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	l.refill(b, now)
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Cleanup удаляет вёдра, которые к текущему моменту пополнились до
// burst: они неотличимы от новых. Возвращает, сколько удалено.
func (l *Limiter) Cleanup() int {
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for k, b := range l.buckets {
		l.refill(b, now)
		if b.tokens >= l.burst {
			delete(l.buckets, k)
			n++
		}
	}
	return n
}

// Len — сколько вёдер сейчас хранится.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
