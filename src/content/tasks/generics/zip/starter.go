package main

type Pair[A, B any] struct {
	First  A
	Second B
}

// Zip склеивает два слайса попарно. Длина результата — по более короткому.
func Zip[A, B any](as []A, bs []B) []Pair[A, B] {
	// ваш код
	return nil
}
