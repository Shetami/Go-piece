package main

func chkBrute(xs []int, k int) []int {
	out := []int{}
	for i := 0; i+k <= len(xs); i++ {
		out = append(out, slices.Max(xs[i:i+k]))
	}
	return out
}

func TestWindowMaxExamples(t *testing.T) {
	cases := []struct {
		xs   []int
		k    int
		want []int
	}{
		{[]int{1, 3, -1, -3, 5, 3, 6, 7}, 3, []int{3, 3, 5, 5, 6, 7}},
		{[]int{4, 4, 4, 1}, 2, []int{4, 4, 4}},
		{[]int{9, 8, 7, 6}, 2, []int{9, 8, 7}},
		{[]int{5, 1, 5, 1, 5}, 2, []int{5, 5, 5, 5}},
		{[]int{2, 7, 1}, 1, []int{2, 7, 1}},
		{[]int{2, 7, 1}, 3, []int{7}},
		{[]int{-5, -2, -9}, 2, []int{-2, -2}},
	}
	for _, c := range cases {
		if got := WindowMax(c.xs, c.k); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("WindowMax(%v, %d) = %v, ожидали %v", c.xs, c.k, got, c.want)
		}
	}
	if got := WindowMax([]int{1, 2}, 3); len(got) != 0 {
		t.Fatalf("окно длиннее входа: %v, ожидали пусто", got)
	}
}

func TestWindowMaxRandom(t *testing.T) {
	seed := uint32(42)
	for trial := range 200 {
		n := trial%30 + 1
		xs := make([]int, n)
		for i := range xs {
			seed = seed*1664525 + 1013904223
			xs[i] = int(seed>>24) % 7 // много повторов
		}
		k := trial%n + 1
		if got, want := WindowMax(xs, k), chkBrute(xs, k); !reflect.DeepEqual(got, want) {
			t.Fatalf("WindowMax(%v, %d) = %v, ожидали %v", xs, k, got, want)
		}
	}
}

func TestWindowMaxLinear(t *testing.T) {
	n := 300000
	xs := make([]int, n)
	for i := range xs {
		xs[i] = n - i // убывание — худший случай для перебора окна
	}
	start := time.Now()
	got := WindowMax(xs, n/2)
	if len(got) != n-n/2+1 || got[0] != n || got[len(got)-1] != n/2 {
		t.Fatalf("неверный результат на убывающем входе")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("n=%d, k=%d заняло %v — окно пересчитывается целиком, это O(n·k)", n, n/2, d)
	}
}

func TestWindowMaxPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("WindowMax(xs, 0): ожидали панику")
		}
	}()
	WindowMax([]int{1}, 0)
}
