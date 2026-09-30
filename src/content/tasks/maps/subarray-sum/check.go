package main

func chkBrute(nums []int, k int) (int, int) {
	c, l := 0, 0
	for i := range nums {
		s := 0
		for j := i; j < len(nums); j++ {
			s += nums[j]
			if s == k {
				c++
				l = max(l, j-i+1)
			}
		}
	}
	return c, l
}

func chkSub(t *testing.T, nums []int, k int) {
	t.Helper()
	wc, wl := chkBrute(nums, k)
	if c, l := SubarraySum(nums, k); c != wc || l != wl {
		t.Fatalf("SubarraySum(%v, %d) = %d, %d; ожидали %d, %d", nums, k, c, l, wc, wl)
	}
}

func TestSubarrayBasic(t *testing.T) {
	chkSub(t, []int{1, 1, 1}, 2)       // 2, 2
	chkSub(t, []int{1, 2, 3, 4, 5}, 9) // [2 3 4] и [4 5]
	chkSub(t, []int{3}, 3)             // весь массив с начала — нужен пустой префикс
	chkSub(t, []int{1, 2}, 3)
}

func TestSubarrayNegativesAndZeros(t *testing.T) {
	chkSub(t, []int{1, -1, 5, -2, 3}, 3)
	chkSub(t, []int{0, 0, 0}, 0) // 6 отрезков, самый длинный 3
	chkSub(t, []int{-2, -1, 2, 1}, 1)
	chkSub(t, []int{2, -2, 2, -2, 2}, 2)
	chkSub(t, []int{5, -5, 5, -5}, 0)
}

func TestSubarrayEdges(t *testing.T) {
	if c, l := SubarraySum(nil, 0); c != 0 || l != 0 {
		t.Fatalf("пустой слайс, k=0: %d, %d; ожидали 0, 0 — пустой отрезок не считается", c, l)
	}
	chkSub(t, []int{1, 2, 3}, 100)
}

func TestSubarrayRandom(t *testing.T) {
	seed := uint32(12345)
	next := func(n int) int { seed = seed*1664525 + 1013904223; return int(seed>>16) % n }
	for range 200 {
		n := next(30)
		nums := make([]int, n)
		for i := range nums {
			nums[i] = next(7) - 3
		}
		chkSub(t, nums, next(7)-3)
	}
}

func TestSubarrayLarge(t *testing.T) {
	nums := make([]int, 200000)
	for i := range nums {
		nums[i] = i%3 - 1 // -1 0 1 -1 0 1 …
	}
	done := make(chan struct{})
	go func() { SubarraySum(nums, 0); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("200 000 элементов не обработаны за разумное время — нужен O(n), а не O(n²)")
	}
}
