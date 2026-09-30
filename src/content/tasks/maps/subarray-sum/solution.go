package main

// SubarraySum ищет непрерывные непустые подотрезки nums с суммой ровно k.
// Возвращает их количество и длину самого длинного из них (0, если таких нет).
// Числа бывают отрицательными и нулевыми. Сложность — O(n).
func SubarraySum(nums []int, k int) (count, longest int) {
	// Сумма отрезка (i, j] = prefix[j] - prefix[i]. Ищем prefix[i] = prefix[j] - k.
	seen := map[int]int{0: 1}   // префиксная сумма → сколько раз встречалась
	first := map[int]int{0: -1} // префиксная сумма → первый индекс (пустой префикс — -1)
	sum := 0
	for j, x := range nums {
		sum += x
		count += seen[sum-k]
		if i, ok := first[sum-k]; ok {
			longest = max(longest, j-i)
		}
		seen[sum]++
		if _, ok := first[sum]; !ok {
			first[sum] = j // только первый: самый ранний даёт самый длинный отрезок
		}
	}
	return count, longest
}
