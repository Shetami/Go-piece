package main

// Reduce сворачивает слайс в одно значение: начиная с init,
// применяет f к накопленному значению и очередному элементу слева направо.
// Тип результата может отличаться от типа элементов.
func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}
