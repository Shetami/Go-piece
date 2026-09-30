package main

// BiMap — взаимно-однозначное соответствие K ↔ V: у каждого ключа не больше
// одного значения, у каждого значения не больше одного ключа.
//
// Put(k, v) связывает k и v. Если k был связан с другим значением или v —
// с другим ключом, эти старые связи исчезают целиком, в обе стороны.
// Delete* удаляют пару с обеих сторон и сообщают, была ли она.
// Нулевые значения ("" и 0) — обычные ключи и значения.
type BiMap[K, V comparable] struct {
	fwd map[K]V
	rev map[V]K
}

func NewBiMap[K, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{fwd: make(map[K]V), rev: make(map[V]K)}
}

func (b *BiMap[K, V]) Put(k K, v V) {
	// Сначала рвём старые связи обоих концов, иначе в обратной мапе
	// останутся «висячие» записи, указывающие на чужую пару.
	if old, ok := b.fwd[k]; ok {
		delete(b.rev, old)
	}
	if old, ok := b.rev[v]; ok {
		delete(b.fwd, old)
	}
	b.fwd[k] = v
	b.rev[v] = k
}

func (b *BiMap[K, V]) GetByKey(k K) (V, bool) {
	v, ok := b.fwd[k]
	return v, ok
}

func (b *BiMap[K, V]) GetByValue(v V) (K, bool) {
	k, ok := b.rev[v]
	return k, ok
}

func (b *BiMap[K, V]) DeleteKey(k K) bool {
	v, ok := b.fwd[k]
	if !ok {
		return false
	}
	delete(b.fwd, k)
	delete(b.rev, v)
	return true
}

func (b *BiMap[K, V]) DeleteValue(v V) bool {
	k, ok := b.rev[v]
	if !ok {
		return false
	}
	delete(b.rev, v)
	delete(b.fwd, k)
	return true
}

// Len — число пар.
func (b *BiMap[K, V]) Len() int { return len(b.fwd) }
