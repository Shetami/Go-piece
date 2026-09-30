package main

import (
	"sync"
	"sync/atomic"
)

// Lazy лениво вычисляет значение через init при первом Get.
//
//   - Успешный результат запоминается навсегда: init больше не вызывается,
//     Get возвращает его без блокировок на мьютексе.
//   - Ошибка НЕ запоминается: следующий Get вызовет init снова.
//   - В каждый момент работает не больше одного вызова init; одновременные
//     Get ждут его. Если он упал, ждавшие пробуют сами — по одному.
//   - Если init запаниковал, паника уходит вызывающему Get, а Lazy остаётся
//     рабочим: следующий Get снова попробует init.
type Lazy[T any] struct {
	init func() (T, error)
	mu   sync.Mutex
	done atomic.Bool // true — val готов; читается без мьютекса
	val  T
}

func NewLazy[T any](init func() (T, error)) *Lazy[T] {
	return &Lazy[T]{init: init}
}

func (l *Lazy[T]) Get() (T, error) {
	// Быстрый путь: после успеха — только атомарное чтение.
	if l.done.Load() {
		return l.val, nil
	}
	l.mu.Lock()
	defer l.mu.Unlock() // defer: паника в init не должна оставить мьютекс занятым
	if l.done.Load() {  // пока ждали мьютекс, кто-то мог уже посчитать
		return l.val, nil
	}
	v, err := l.init()
	if err != nil {
		var zero T
		return zero, err
	}
	l.val = v
	l.done.Store(true) // Store после записи val — публикация значения
	return v, nil
}
