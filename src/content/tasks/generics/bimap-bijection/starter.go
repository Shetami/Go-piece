package main

// BiMap — взаимно однозначное соответствие: у каждого ключа одно
// значение, у каждого значения один ключ. Создаётся через NewBiMap.
type BiMap[K, V comparable] struct {
	// ваши поля
}

func NewBiMap[K, V comparable]() *BiMap[K, V] {
	// ваш код
	return &BiMap[K, V]{}
}

// Put связывает k и v. Прежние связи k и v, если были, удаляются
// с обеих сторон: после Put(k, v) у k только значение v, у v только ключ k.
func (m *BiMap[K, V]) Put(k K, v V) {
	// ваш код
}

// Get возвращает значение ключа.
func (m *BiMap[K, V]) Get(k K) (V, bool) {
	// ваш код
	var zero V
	return zero, false
}

// GetKey возвращает ключ значения.
func (m *BiMap[K, V]) GetKey(v V) (K, bool) {
	// ваш код
	var zero K
	return zero, false
}

// Delete удаляет ключ вместе с его значением; false, если ключа не было.
func (m *BiMap[K, V]) Delete(k K) bool {
	// ваш код
	return false
}

// Len возвращает число пар.
func (m *BiMap[K, V]) Len() int {
	// ваш код
	return 0
}

// Inverse возвращает ту же биекцию с переставленными ролями. Это
// представление, а не копия: изменения через него видны в m и наоборот.
func (m *BiMap[K, V]) Inverse() *BiMap[V, K] {
	// ваш код
	return &BiMap[V, K]{}
}
