package main

// Mailbox — буфер на capacity последних значений (capacity >= 1).
// Push никогда не блокируется: если буфер полон, самое старое значение
// выбрасывается. Подходит для телеметрии и состояния, где свежие данные
// важнее полноты. Безопасен для многих писателей и читателей.
type Mailbox[T any] struct {
	// ваши поля
}

func NewMailbox[T any](capacity int) *Mailbox[T] {
	// ваш код
	return &Mailbox[T]{}
}

// Push кладёт v. Возвращает true, если ради этого выброшено старое значение.
func (m *Mailbox[T]) Push(v T) (dropped bool) {
	// ваш код
	return false
}

// C — канал, из которого читают значения. Каждый вызов — один и тот же канал.
func (m *Mailbox[T]) C() <-chan T {
	// ваш код
	return nil
}

// Dropped — сколько значений выброшено за всё время.
func (m *Mailbox[T]) Dropped() int {
	// ваш код
	return 0
}
