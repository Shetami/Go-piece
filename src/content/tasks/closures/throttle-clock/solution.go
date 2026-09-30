package main

import (
	"sync"
	"time"
)

// NewThrottle пропускает f не чаще раза в interval. Время сообщают call и tick.
//
// call(v, now): если f ещё не вызывалась или с последнего вызова f прошло
// не меньше interval — f(v) вызывается сразу. Иначе v откладывается
// (последнее значение побеждает).
// tick(now): если есть отложенное значение и с последнего вызова f прошло
// не меньше interval — f вызывается с ним. Это тоже вызов f: следующий
// интервал отсчитывается от него.
// Если call пришёл, когда интервал уже истёк, а отложенное значение так и
// не отправлено, оно устарело: f вызывается один раз, с новым v.
//
// call и tick безопасны для горутин. f вызывается не под блокировкой:
// она может сама вызывать call.
func NewThrottle[T any](interval time.Duration, f func(T)) (call func(v T, now time.Time), tick func(now time.Time)) {
	var (
		mu      sync.Mutex
		fired   bool // f уже вызывалась
		last    time.Time
		pending bool
		pval    T
	)
	// ready — можно ли звать f прямо сейчас; вызывать под mu.
	ready := func(now time.Time) bool {
		return !fired || now.Sub(last) >= interval
	}
	// fire отмечает вызов под замком; сам вызов f — после Unlock.
	fire := func(now time.Time) {
		fired, last = true, now
		pending = false
		var zero T
		pval = zero
	}
	call = func(v T, now time.Time) {
		mu.Lock()
		if !ready(now) {
			pending, pval = true, v
			mu.Unlock()
			return
		}
		fire(now) // отложенное, если было, устарело и выбрасывается
		mu.Unlock()
		f(v)
	}
	tick = func(now time.Time) {
		mu.Lock()
		if !pending || !ready(now) {
			mu.Unlock()
			return
		}
		v := pval
		fire(now)
		mu.Unlock()
		f(v)
	}
	return call, tick
}
