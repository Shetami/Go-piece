package main

// MemoRec мемоизирует рекурсивную функцию. f получает self — функцию,
// через которую делаются рекурсивные вызовы; они тоже идут через кэш.
// Каждое значение k вычисляется (f вызывается) не больше одного раза за
// жизнь результата. У каждого вызова MemoRec свой кэш. Горутины — не нужны.
func MemoRec[K comparable, V any](f func(self func(K) V, k K) V) func(K) V {
	// ваш код
	return func(k K) V {
		var zero V
		return zero
	}
}
