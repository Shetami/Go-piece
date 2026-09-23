package main

// Memoize возвращает функцию, которая считает f(k) один раз на каждый k,
// а дальше отдаёт запомненный результат.
func Memoize[K comparable, V any](f func(K) V) func(K) V {
	// Кэш живёт в замыкании: он создаётся один раз, при вызове Memoize,
	// и свой у каждой мемоизированной функции.
	cache := make(map[K]V)
	return func(k K) V {
		if v, ok := cache[k]; ok {
			return v
		}
		v := f(k)
		cache[k] = v
		return v
	}
}
