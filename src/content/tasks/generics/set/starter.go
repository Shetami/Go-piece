package main

// Set — множество значений типа T.
type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	// ваш код
	return nil
}

func (s Set[T]) Add(v T) {
	// ваш код
}

func (s Set[T]) Has(v T) bool {
	// ваш код
	return false
}

// Union и Intersect возвращают новое множество и не меняют исходные.
func (s Set[T]) Union(o Set[T]) Set[T] {
	// ваш код
	return nil
}

func (s Set[T]) Intersect(o Set[T]) Set[T] {
	// ваш код
	return nil
}
