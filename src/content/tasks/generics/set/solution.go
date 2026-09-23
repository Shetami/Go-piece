package main

// Set — множество значений типа T.
type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, v := range items {
		s.Add(v)
	}
	return s
}

// Методы на значении: Set — мапа, а мапа и так ссылается на общие данные.
func (s Set[T]) Add(v T) { s[v] = struct{}{} }

func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

// Union и Intersect возвращают новое множество и не меняют исходные.
func (s Set[T]) Union(o Set[T]) Set[T] {
	out := make(Set[T], len(s)+len(o))
	for v := range s {
		out.Add(v)
	}
	for v := range o {
		out.Add(v)
	}
	return out
}

func (s Set[T]) Intersect(o Set[T]) Set[T] {
	// Обходим меньшее множество, проверяем в большем.
	small, big := s, o
	if len(big) < len(small) {
		small, big = big, small
	}
	out := make(Set[T])
	for v := range small {
		if big.Has(v) {
			out.Add(v)
		}
	}
	return out
}
