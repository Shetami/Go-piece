package main

// ForEachLimit вызывает f для каждого элемента, одновременно — не больше limit
// вызовов. Возвращается, когда все вызовы закончились.
func ForEachLimit(items []int, limit int, f func(int)) {
	// ваш код
	for _, it := range items {
		f(it)
	}
}
