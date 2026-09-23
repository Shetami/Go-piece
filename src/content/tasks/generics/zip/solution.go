package main

type Pair[A, B any] struct {
	First  A
	Second B
}

// Zip склеивает два слайса попарно. Длина результата — по более короткому.
func Zip[A, B any](as []A, bs []B) []Pair[A, B] {
	n := min(len(as), len(bs))
	out := make([]Pair[A, B], n)
	for i := range n {
		out[i] = Pair[A, B]{as[i], bs[i]}
	}
	return out
}
