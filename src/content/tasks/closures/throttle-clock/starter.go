package main

import "time"

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
	// ваш код
	return func(v T, now time.Time) { f(v) }, func(time.Time) {}
}
