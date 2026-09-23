package main

// Pipe собирает функции в одну: Pipe(f, g, h)(x) == h(g(f(x))).
// Без аргументов — тождественная функция.
func Pipe(fs ...func(int) int) func(int) int {
	// Копия: вариативный параметр, переданный как fs..., делит массив
	// с вызывающим, и тот может поменять функции уже после сборки.
	fs = append([]func(int) int(nil), fs...)
	return func(x int) int {
		for _, f := range fs {
			x = f(x)
		}
		return x
	}
}
