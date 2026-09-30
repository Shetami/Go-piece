package main

import "sync"

// Mailbox — буфер на capacity последних значений (capacity >= 1).
// Push никогда не блокируется: если буфер полон, самое старое значение
// выбрасывается. Подходит для телеметрии и состояния, где свежие данные
// важнее полноты. Безопасен для многих писателей и читателей.
type Mailbox[T any] struct {
	ch chan T

	mu      sync.Mutex // сериализует писателей: «выбросить и положить» — одна операция
	dropped int
}

func NewMailbox[T any](capacity int) *Mailbox[T] {
	return &Mailbox[T]{ch: make(chan T, capacity)}
}

// Push кладёт v. Возвращает true, если ради этого выброшено старое значение.
func (m *Mailbox[T]) Push(v T) (dropped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for {
		select {
		case m.ch <- v:
			return dropped
		default:
		}
		// Буфер полон. Выбрасываем старейшее — тоже без блокировки:
		// читатель мог успеть его забрать, тогда просто пробуем снова.
		select {
		case <-m.ch:
			dropped = true
			m.dropped++
		default:
		}
	}
}

// C — канал, из которого читают значения. Каждый вызов — один и тот же канал.
func (m *Mailbox[T]) C() <-chan T {
	return m.ch
}

// Dropped — сколько значений выброшено за всё время.
func (m *Mailbox[T]) Dropped() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dropped
}
