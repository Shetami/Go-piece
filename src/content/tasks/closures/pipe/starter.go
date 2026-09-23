package main

// Pipe собирает функции в одну: Pipe(f, g, h)(x) == h(g(f(x))).
// Без аргументов — тождественная функция.
func Pipe(fs ...func(int) int) func(int) int {
	// ваш код
	return nil
}
