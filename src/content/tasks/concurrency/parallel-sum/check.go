package main

func checkNums(n int) ([]int, int) {
	nums := make([]int, n)
	total := 0
	for i := range nums {
		nums[i] = i*7%13 - 3
		total += nums[i]
	}
	return nums, total
}

func TestParallelSumBasic(t *testing.T) {
	nums, want := checkNums(1000)
	for _, parts := range []int{1, 2, 3, 7, 10, 999, 1000} {
		if got := ParallelSum(nums, parts); got != want {
			t.Fatalf("ParallelSum(1000 чисел, %d частей) = %d, ожидали %d", parts, got, want)
		}
	}
}

func TestParallelSumOddSplit(t *testing.T) {
	if got := ParallelSum([]int{1, 2, 3, 4, 5}, 2); got != 15 {
		t.Fatalf("5 чисел на 2 части: %d, ожидали 15 — последний хвост потерялся?", got)
	}
}

func TestParallelSumEdge(t *testing.T) {
	if got := ParallelSum([]int{1, 2, 3}, 10); got != 6 {
		t.Fatalf("частей больше, чем чисел: %d, ожидали 6", got)
	}
	if got := ParallelSum([]int{4, 5}, 0); got != 9 {
		t.Fatalf("parts = 0: %d, ожидали 9", got)
	}
	if got := ParallelSum(nil, 4); got != 0 {
		t.Fatalf("пустой вход: %d", got)
	}
}
