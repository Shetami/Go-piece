package main

import (
	"sync"
	"sync/atomic"
)

// Lazy лениво вычисляет значение через init.
type Lazy[T any] struct {
	init func() (T, error)
	mu   sync.Mutex
	done atomic.Bool
	val  T
}

func NewLazy[T any](init func() (T, error)) *Lazy[T] {
	return &Lazy[T]{init: init}
}

// Get возвращает значение, вычисляя его при первом успешном вызове init.
//   - Значение уже есть — вернуть его, init не вызывать.
//   - Одновременно выполняется не больше одного init; остальные Get ждут.
//   - init вернула ошибку — вернуть её; значение не запоминается, следующий
//     Get вызовет init снова.
//   - init запаниковала — паника летит из этого Get дальше, а Lazy остаётся
//     рабочим: следующий Get снова вызовет init (и не зависнет).
//
// Безопасен для конкурентного использования.
func (l *Lazy[T]) Get() (T, error) {
	// Быстрый путь без блокировки. val записан до done.Store(true),
	// поэтому, увидев true, читать val безопасно.
	if l.done.Load() {
		return l.val, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock() // отпустит мьютекс и при панике init
	if l.done.Load() {  // пока ждали, кто-то мог уже вычислить
		return l.val, nil
	}
	v, err := l.init()
	if err != nil {
		var zero T
		return zero, err
	}
	l.val = v
	l.done.Store(true) // только после успеха: ошибка и паника не «засчитываются»
	return v, nil
}
