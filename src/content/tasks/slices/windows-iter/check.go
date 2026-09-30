package main

func chkCollect[T any](seq iter.Seq[[]T]) [][]T {
	out := [][]T{}
	for w := range seq {
		out = append(out, w)
	}
	return out
}

func TestWindowsShapes(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6, 7}
	cases := []struct {
		size, step int
		want       [][]int
	}{
		{3, 1, [][]int{{1, 2, 3}, {2, 3, 4}, {3, 4, 5}, {4, 5, 6}, {5, 6, 7}}},
		{3, 2, [][]int{{1, 2, 3}, {3, 4, 5}, {5, 6, 7}}},
		{2, 3, [][]int{{1, 2}, {4, 5}}},
		{7, 1, [][]int{{1, 2, 3, 4, 5, 6, 7}}},
		{8, 1, [][]int{}},
		{1, 3, [][]int{{1}, {4}, {7}}},
	}
	for _, c := range cases {
		if got := chkCollect(Windows(s, c.size, c.step)); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("Windows(%v, size=%d, step=%d) = %v, ожидали %v", s, c.size, c.step, got, c.want)
		}
	}
	seq := Windows([]string{"a", "b", "c"}, 2, 1)
	if a, b := chkCollect(seq), chkCollect(seq); !reflect.DeepEqual(a, b) || len(a) != 2 {
		t.Fatalf("повторный обход дал %v, первый — %v", b, a)
	}
}

func TestWindowsBreak(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("после break итератор продолжил звать yield: %v", r)
		}
	}()
	n := 0
	for range Windows(make([]int, 100), 2, 1) {
		n++
		if n == 3 {
			break
		}
	}
	if n != 3 {
		t.Fatalf("до break прошли %d окон", n)
	}
}

func TestWindowsAppendSafe(t *testing.T) {
	s := []int{1, 2, 3, 4}
	for w := range Windows(s, 2, 2) {
		_ = append(w, 100)
	}
	if !reflect.DeepEqual(s, []int{1, 2, 3, 4}) {
		t.Fatalf("append к окну изменил исходный слайс: %v", s)
	}
	var first []int
	for w := range Windows(s, 2, 2) {
		first = w
		break
	}
	if &first[0] != &s[0] {
		t.Fatalf("окно — копия, а должно быть подслайсом s")
	}
}

func TestWindowsNoAllocPerWindow(t *testing.T) {
	s := make([]int, 1000)
	allocs := testing.AllocsPerRun(20, func() {
		for w := range Windows(s, 10, 1) {
			_ = w
		}
	})
	if allocs > 2 {
		t.Fatalf("обход 991 окна делает %.0f аллокаций — окна копируются?", allocs)
	}
}

func TestWindowsPanicsEagerly(t *testing.T) {
	for _, c := range [][2]int{{0, 1}, {1, 0}, {-1, 1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("Windows(size=%d, step=%d) без обхода: ожидали панику сразу", c[0], c[1])
				}
			}()
			_ = Windows([]int{1, 2}, c[0], c[1])
		}()
	}
}
