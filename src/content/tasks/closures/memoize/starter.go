package main

// Memoize возвращает функцию, которая считает f(k) один раз на каждый k,
// а дальше отдаёт запомненный результат.
func Memoize[K comparable, V any](f func(K) V) func(K) V {
	// ваш код
	return f
}
