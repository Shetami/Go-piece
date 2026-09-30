package main

// BiMap — взаимно однозначное соответствие: у каждого ключа одно
// значение, у каждого значения один ключ. Создаётся через NewBiMap.
type BiMap[K, V comparable] struct {
	fwd map[K]V
	rev map[V]K
}

func NewBiMap[K, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{fwd: map[K]V{}, rev: map[V]K{}}
}

// Put связывает k и v. Прежние связи k и v, если были, удаляются
// с обеих сторон: после Put(k, v) у k только значение v, у v только ключ k.
func (m *BiMap[K, V]) Put(k K, v V) {
	if old, ok := m.fwd[k]; ok {
		delete(m.rev, old) // старое значение k больше никому не принадлежит
	}
	if oldK, ok := m.rev[v]; ok {
		delete(m.fwd, oldK) // v отбираем у прежнего ключа
	}
	m.fwd[k] = v
	m.rev[v] = k
}

// Get возвращает значение ключа.
func (m *BiMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.fwd[k]
	return v, ok
}

// GetKey возвращает ключ значения.
func (m *BiMap[K, V]) GetKey(v V) (K, bool) {
	k, ok := m.rev[v]
	return k, ok
}

// Delete удаляет ключ вместе с его значением; false, если ключа не было.
func (m *BiMap[K, V]) Delete(k K) bool {
	v, ok := m.fwd[k]
	if !ok {
		return false
	}
	delete(m.fwd, k)
	delete(m.rev, v)
	return true
}

// Len возвращает число пар.
func (m *BiMap[K, V]) Len() int { return len(m.fwd) }

// Inverse возвращает ту же биекцию с переставленными ролями. Это
// представление, а не копия: изменения через него видны в m и наоборот.
func (m *BiMap[K, V]) Inverse() *BiMap[V, K] {
	return &BiMap[V, K]{fwd: m.rev, rev: m.fwd} // те же мапы, типы переставлены
}
