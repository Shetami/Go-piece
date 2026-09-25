package main

// TwoSum возвращает индексы i < j двух разных элементов, сумма которых равна
// target, и true. Если подходящих пар несколько — любую. Если нет — false.
// Сложность — O(n).
func TwoSum(nums []int, target int) (int, int, bool) {
	// Значение → индекс, где оно уже встречалось.
	seen := make(map[int]int, len(nums))
	for j, x := range nums {
		// Сначала ищем пару среди прошлых, потом запоминаем текущий:
		// так элемент не сложится сам с собой.
		if i, ok := seen[target-x]; ok {
			return i, j, true
		}
		seen[x] = j
	}
	return 0, 0, false
}
