package main

func checkRotate(t *testing.T, in []int, k int, want []int) {
	t.Helper()
	xs := append([]int(nil), in...)
	Rotate(xs, k)
	if !reflect.DeepEqual(xs, want) {
		t.Fatalf("Rotate(%v, %d) дал %v, ожидали %v", in, k, xs, want)
	}
}

func TestRotateLeft(t *testing.T) {
	checkRotate(t, []int{1, 2, 3, 4, 5}, 2, []int{3, 4, 5, 1, 2})
	checkRotate(t, []int{1, 2, 3, 4, 5}, 0, []int{1, 2, 3, 4, 5})
}

func TestRotateBigK(t *testing.T) {
	checkRotate(t, []int{1, 2, 3}, 7, []int{2, 3, 1})
	checkRotate(t, []int{1, 2, 3}, 3, []int{1, 2, 3})
}

func TestRotateNegative(t *testing.T) {
	checkRotate(t, []int{1, 2, 3, 4, 5}, -1, []int{5, 1, 2, 3, 4})
	checkRotate(t, []int{1, 2, 3}, -4, []int{3, 1, 2})
}

func TestRotateEmpty(t *testing.T) {
	Rotate(nil, 3)
	Rotate([]int{}, -2)
}
