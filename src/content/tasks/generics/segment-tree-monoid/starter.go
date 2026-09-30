package main

// SegTree — дерево отрезков над моноидом: combine ассоциативна
// (но не обязательно коммутативна), identity — её нейтральный элемент.
type SegTree[T any] struct {
	// ваши поля
}

// NewSegTree строит дерево по vals за O(n). vals не запоминается:
// последующие изменения vals на дерево не влияют.
func NewSegTree[T any](vals []T, identity T, combine func(a, b T) T) *SegTree[T] {
	// ваш код
	return &SegTree[T]{}
}

// Set делает элемент i равным v за O(log n).
func (s *SegTree[T]) Set(i int, v T) {
	// ваш код
}

// Query возвращает combine(v[l], v[l+1], …, v[r-1]) строго слева
// направо; для l >= r — identity. O(log n).
func (s *SegTree[T]) Query(l, r int) T {
	// ваш код
	var zero T
	return zero
}
