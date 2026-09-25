package main

// Chunk режет s на куски по n элементов; последний кусок может быть короче.
// Куски независимы: append к одному не должен менять соседний.
// Для n <= 0 — паника.
func Chunk[T any](s []T, n int) [][]T {
	if n <= 0 {
		panic("Chunk: n должно быть больше нуля")
	}
	out := make([][]T, 0, (len(s)+n-1)/n)
	for len(s) > 0 {
		end := min(n, len(s))
		// Третий индекс обрезает вместимость: append к куску уйдёт в новый массив.
		out = append(out, s[:end:end])
		s = s[end:]
	}
	return out
}
