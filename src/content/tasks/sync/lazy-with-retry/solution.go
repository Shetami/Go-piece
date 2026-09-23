package main

import "sync"

// Lazy лениво вычисляет значение. В отличие от sync.Once,
// неудачная инициализация не запоминается: следующий Get попробует снова.
// После первого успеха init больше не вызывается.
// Одновременно выполняется не больше одного init.
type Lazy[T any] struct {
	mu    sync.Mutex
	done  bool
	value T
}

func (l *Lazy[T]) Get(init func() (T, error)) (T, error) {
	// Мьютекс держим и во время init: остальные подождут результата,
	// а не запустят свою инициализацию параллельно.
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done {
		return l.value, nil
	}
	v, err := init()
	if err != nil {
		var zero T
		return zero, err
	}
	l.value, l.done = v, true
	return v, nil
}
