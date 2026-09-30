package main

import "time"

// TTLCache — потокобезопасный кэш, у каждой записи свой срок жизни.
// Время берётся только из now — так кэш можно тестировать без sleep.
//
//   - Set кладёт значение со сроком ttl; ttl <= 0 — бессрочно. Повторный Set
//     того же ключа заменяет и значение, и срок.
//   - Запись жива, пока now() строго раньше момента истечения. Get истёкшей
//     записи — промах, запись при этом удаляется.
//   - Len — число живых записей на текущий момент.
//   - Purge удаляет все истёкшие записи и возвращает их число. Он не должен
//     обходить все записи: храните сроки в куче (container/heap).
type TTLCache[K comparable, V any] struct {
	// ваши поля
}

func NewTTLCache[K comparable, V any](now func() time.Time) *TTLCache[K, V] {
	return &TTLCache[K, V]{}
}

func (c *TTLCache[K, V]) Set(k K, v V, ttl time.Duration) {
	// ваш код
}

func (c *TTLCache[K, V]) Get(k K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

func (c *TTLCache[K, V]) Len() int {
	// ваш код
	return 0
}

func (c *TTLCache[K, V]) Purge() int {
	// ваш код
	return 0
}
