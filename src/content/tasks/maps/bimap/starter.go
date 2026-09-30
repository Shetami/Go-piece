package main

// BiMap — взаимно-однозначное соответствие K ↔ V: у каждого ключа не больше
// одного значения, у каждого значения не больше одного ключа.
//
// Put(k, v) связывает k и v. Если k был связан с другим значением или v —
// с другим ключом, эти старые связи исчезают целиком, в обе стороны.
// Delete* удаляют пару с обеих сторон и сообщают, была ли она.
// Нулевые значения ("" и 0) — обычные ключи и значения.
type BiMap[K, V comparable] struct {
	// ваши поля
}

func NewBiMap[K, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{}
}

func (b *BiMap[K, V]) Put(k K, v V) {
	// ваш код
}

func (b *BiMap[K, V]) GetByKey(k K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

func (b *BiMap[K, V]) GetByValue(v V) (K, bool) {
	// ваш код
	var zero K
	return zero, false
}

func (b *BiMap[K, V]) DeleteKey(k K) bool {
	// ваш код
	return false
}

func (b *BiMap[K, V]) DeleteValue(v V) bool {
	// ваш код
	return false
}

// Len — число пар.
func (b *BiMap[K, V]) Len() int {
	// ваш код
	return 0
}
