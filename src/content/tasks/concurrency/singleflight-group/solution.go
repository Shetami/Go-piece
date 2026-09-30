package main

import (
	"errors"
	"sync"
)

// ErrPanicked получают ожидающие вызовы Do, если fn запаниковала.
var ErrPanicked = errors.New("singleflight: fn panicked")

// Group объединяет одновременные вызовы с одинаковым ключом.
// Нулевое значение готово к работе.
type Group[K comparable, V any] struct {
	mu    sync.Mutex
	calls map[K]*call[V]
}

type call[V any] struct {
	done   chan struct{} // закрывается, когда val/err готовы
	val    V
	err    error
	shared bool // true, если к вызову кто-то присоединился
}

// Do выполняет fn для key и возвращает её результат.
//
// Если для того же key вызов уже идёт, Do не запускает fn повторно, а ждёт
// текущий и возвращает его результат (и ошибку тоже). shared = true у всех
// участников вызова, если участников было больше одного.
// После завершения fn ключ забывается: следующий Do запустит fn заново.
// Вызовы с разными ключами друг друга не ждут.
// Если fn паникует, паника продолжается в горутине, которая её запускала,
// ожидающие получают ErrPanicked, а ключ освобождается.
func (g *Group[K, V]) Do(key K, fn func() (V, error)) (v V, err error, shared bool) {
	g.mu.Lock()
	if g.calls == nil {
		g.calls = make(map[K]*call[V])
	}
	if c, ok := g.calls[key]; ok {
		c.shared = true
		g.mu.Unlock()
		<-c.done // ждём без мьютекса: другие ключи не блокируются
		return c.val, c.err, true
	}
	c := &call[V]{done: make(chan struct{})}
	c.err = ErrPanicked // перезапишется, если fn вернётся нормально
	g.calls[key] = c
	g.mu.Unlock()

	// defer сработает и при панике: ключ освободится, ждущие проснутся.
	defer func() {
		g.mu.Lock()
		delete(g.calls, key)
		shared = c.shared
		g.mu.Unlock()
		close(c.done)
	}()
	c.val, c.err = fn()
	return c.val, c.err, false // shared выставит defer
}
