package main

// SegTree — дерево отрезков над моноидом: combine ассоциативна
// (но не обязательно коммутативна), identity — её нейтральный элемент.
type SegTree[T any] struct {
	n       int
	t       []T // t[n+i] — листья, t[i] = combine(t[2i], t[2i+1])
	id      T
	combine func(a, b T) T
}

// NewSegTree строит дерево по vals за O(n). vals не запоминается:
// последующие изменения vals на дерево не влияют.
func NewSegTree[T any](vals []T, identity T, combine func(a, b T) T) *SegTree[T] {
	n := len(vals)
	s := &SegTree[T]{n: n, t: make([]T, 2*n), id: identity, combine: combine}
	copy(s.t[n:], vals) // копия, а не срез vals
	for i := n - 1; i > 0; i-- {
		s.t[i] = combine(s.t[2*i], s.t[2*i+1])
	}
	return s
}

// Set делает элемент i равным v за O(log n).
func (s *SegTree[T]) Set(i int, v T) {
	i += s.n
	s.t[i] = v
	for i >>= 1; i > 0; i >>= 1 {
		s.t[i] = s.combine(s.t[2*i], s.t[2*i+1])
	}
}

// Query возвращает combine(v[l], v[l+1], …, v[r-1]) строго слева
// направо; для l >= r — identity. O(log n).
func (s *SegTree[T]) Query(l, r int) T {
	if l >= r {
		return s.id
	}
	// Два аккумулятора: левый растёт вправо, правый — влево. Сложить
	// всё в один нельзя — для некоммутативной combine порядок сломается.
	left, right := s.id, s.id
	for l, r = l+s.n, r+s.n; l < r; l, r = l>>1, r>>1 {
		if l&1 == 1 {
			left = s.combine(left, s.t[l])
			l++
		}
		if r&1 == 1 {
			r--
			right = s.combine(s.t[r], right)
		}
	}
	return s.combine(left, right)
}
