package main

import "time"

// NewDebouncer откладывает f, пока вызовы не утихнут. Время не идёт само:
// его сообщают call и tick (тесты двигают его вручную).
//
// call(v, now) запоминает v как последнее значение серии.
// tick(now) вызывает f с последним значением серии, если:
//   - с последнего call прошло не меньше wait, или
//   - maxWait > 0 и с первого call серии прошло не меньше maxWait.
//
// После вызова f серия закончена, следующий call начинает новую.
// Без отложенного значения tick ничего не делает. Вызовы — из одной горутины.
func NewDebouncer[T any](wait, maxWait time.Duration, f func(T)) (call func(v T, now time.Time), tick func(now time.Time)) {
	var (
		pending     bool // есть ли отложенное значение; сам v может быть нулевым
		last        T
		first, seen time.Time
	)
	call = func(v T, now time.Time) {
		if !pending {
			pending = true
			first = now // начало серии — для maxWait
		}
		last, seen = v, now
	}
	tick = func(now time.Time) {
		if !pending {
			return
		}
		quiet := now.Sub(seen) >= wait
		tooLong := maxWait > 0 && now.Sub(first) >= maxWait
		if !quiet && !tooLong {
			return
		}
		v := last
		pending = false
		var zero T
		last = zero // не держим ссылку на значение после отправки
		f(v)
	}
	return call, tick
}
