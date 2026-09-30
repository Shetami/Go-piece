package main

import "time"

// NewWindowCounter считает события за скользящее окно длиной window,
// разбитое на buckets корзин по size = window/buckets (делится нацело).
// Корзина момента t — id = floor(t.UnixNano() / size).
//
// add(n) кладёт n событий в корзину момента now(). count() возвращает
// сумму по корзинам с id в (cur-buckets, cur], где cur — корзина now().
// Время может идти назад: событие, чья корзина ещё в окне, учитывается;
// событие из корзины, которую уже вытеснила более новая, отбрасывается.
// Память — O(buckets), независимо от числа событий. Безопасно для горутин.
func NewWindowCounter(window time.Duration, buckets int, now func() time.Time) (add func(n int), count func() int) {
	// ваш код
	return func(int) {}, func() int { return 0 }
}
