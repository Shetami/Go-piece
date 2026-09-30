package main

// LFU — кэш на capacity записей с вытеснением наименее часто используемой.
//
//   - Частота записи — сколько раз к ней обращались: Put нового ключа даёт 1,
//     каждый Get-попадание и Put существующего ключа прибавляют 1.
//   - При переполнении вытесняется запись с наименьшей частотой, а среди
//     равных — та, к которой дольше всего не обращались.
//   - Вытесненный и вставленный заново ключ начинает с частоты 1.
//   - capacity <= 0 — кэш ничего не хранит.
//   - Get и Put — за O(1).
type LFU[K comparable, V any] struct {
	// ваши поля
}

func NewLFU[K comparable, V any](capacity int) *LFU[K, V] {
	return &LFU[K, V]{}
}

func (c *LFU[K, V]) Get(k K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

func (c *LFU[K, V]) Put(k K, v V) {
	// ваш код
}

func (c *LFU[K, V]) Len() int {
	// ваш код
	return 0
}
