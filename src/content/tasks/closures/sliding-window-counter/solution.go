package main

import (
	"sync"
	"time"
)

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
	size := int64(window) / int64(buckets)
	type slot struct {
		id    int64 // какая корзина сейчас лежит в ячейке
		n     int
		valid bool
	}
	var (
		mu   sync.Mutex
		ring = make([]slot, buckets)
	)
	bucketOf := func(t time.Time) int64 {
		ns := t.UnixNano()
		id := ns / size
		if ns%size < 0 { // floor для отрицательных
			id--
		}
		return id
	}
	index := func(id int64) int {
		i := int(id % int64(buckets))
		if i < 0 {
			i += buckets
		}
		return i
	}
	add = func(n int) {
		mu.Lock()
		defer mu.Unlock()
		id := bucketOf(now())
		s := &ring[index(id)]
		if !s.valid || s.id != id {
			if s.valid && s.id > id {
				return // ячейку уже заняла более новая корзина — событие устарело
			}
			// Ячейку занимает старая корзина: сначала обнулить, потом писать.
			*s = slot{id: id, valid: true}
		}
		s.n += n
	}
	count = func() int {
		mu.Lock()
		defer mu.Unlock()
		cur := bucketOf(now())
		total := 0
		for _, s := range ring {
			// Проверяем id, а не доверяем ячейке: после простоя в ней
			// лежат корзины давно ушедших окон.
			if s.valid && s.id > cur-int64(buckets) && s.id <= cur {
				total += s.n
			}
		}
		return total
	}
	return add, count
}
