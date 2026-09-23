package main

import "sync"

// ParallelSum складывает числа, разбив слайс на parts частей и посчитав
// каждую часть в своей горутине. parts < 1 считается как 1;
// частей больше, чем чисел, быть не должно.
func ParallelSum(nums []int, parts int) int {
	parts = max(1, min(parts, len(nums)))
	if len(nums) == 0 {
		return 0
	}
	// Каждая горутина пишет в свою ячейку — мьютекс не нужен.
	sums := make([]int, parts)
	size := (len(nums) + parts - 1) / parts // округление вверх
	var wg sync.WaitGroup
	for p := range parts {
		lo := min(p*size, len(nums)) // хвостовые части могут оказаться пустыми
		hi := min(lo+size, len(nums))
		wg.Go(func() {
			for _, n := range nums[lo:hi] {
				sums[p] += n
			}
		})
	}
	wg.Wait()

	total := 0
	for _, s := range sums {
		total += s
	}
	return total
}
