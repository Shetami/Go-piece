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
	// ваш код
	return func(T, time.Time) {}, func(time.Time) {}
}
