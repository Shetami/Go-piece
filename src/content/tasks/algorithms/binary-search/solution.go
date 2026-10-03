package main

// Search ищет value в отсортированном по возрастанию слайсе и возвращает
// его индекс или -1, если такого значения нет.
func Search(array []int, value int) int {
	// Полуинтервал [lo, hi): искомое, если есть, лежит в нём.
	lo, hi := 0, len(array)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		switch {
		case array[mid] == value:
			return mid
		case array[mid] < value:
			lo = mid + 1
		default:
			hi = mid
		}
	}
	return -1
}
