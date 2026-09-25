package main

// LowerBound возвращает индекс первого элемента отсортированного по возрастанию
// слайса, который >= x. Если таких нет — len(s).
// Сложность — O(log n). Функции пакетов sort и slices не использовать.
func LowerBound(s []int, x int) int {
	// Инвариант: всё левее lo — меньше x, всё начиная с hi — не меньше x.
	lo, hi := 0, len(s)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1) // без переполнения при огромных индексах
		if s[mid] < x {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}
