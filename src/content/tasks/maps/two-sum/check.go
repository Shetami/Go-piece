package main

func checkPair(t *testing.T, nums []int, target int) {
	t.Helper()
	i, j, ok := TwoSum(nums, target)
	if !ok {
		t.Fatalf("TwoSum(%v, %d): пара есть, а вернули false", nums, target)
	}
	if !(0 <= i && i < j && j < len(nums)) || nums[i]+nums[j] != target {
		t.Fatalf("TwoSum(%v, %d) = (%d, %d): неверная пара", nums, target, i, j)
	}
}

func TestTwoSumBasic(t *testing.T) {
	checkPair(t, []int{2, 7, 11, 15}, 9)
	checkPair(t, []int{3, 2, 4}, 6)
	checkPair(t, []int{-3, 4, 3, 90}, 0)
}

func TestTwoSumSameValue(t *testing.T) {
	checkPair(t, []int{3, 3}, 6)
}

func TestTwoSumNotSelf(t *testing.T) {
	if i, j, ok := TwoSum([]int{3, 5}, 6); ok {
		t.Fatalf("TwoSum([3 5], 6) = (%d, %d, true): элемент нельзя складывать сам с собой", i, j)
	}
}

func TestTwoSumNone(t *testing.T) {
	if _, _, ok := TwoSum([]int{1, 2, 3}, 100); ok {
		t.Fatalf("пары нет, а вернули true")
	}
	if _, _, ok := TwoSum(nil, 0); ok {
		t.Fatalf("пустой вход, а вернули true")
	}
}

func TestTwoSumLinear(t *testing.T) {
	n := 200000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i * 2
	}
	start := time.Now()
	checkPair(t, nums, nums[n-1]+nums[n-2])
	if d := time.Since(start); d > time.Second {
		t.Fatalf("200 000 элементов заняли %v — похоже на O(n²)", d)
	}
}
