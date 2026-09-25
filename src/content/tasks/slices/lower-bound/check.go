package main

func TestLowerBoundCases(t *testing.T) {
	s := []int{1, 3, 3, 3, 7, 9}
	cases := []struct{ x, want int }{
		{0, 0}, {1, 0}, {2, 1}, {3, 1}, {4, 4}, {7, 4}, {8, 5}, {9, 5}, {10, 6},
	}
	for _, c := range cases {
		if got := LowerBound(s, c.x); got != c.want {
			t.Fatalf("LowerBound(%v, %d) = %d, ожидали %d", s, c.x, got, c.want)
		}
	}
}

func TestLowerBoundEmpty(t *testing.T) {
	if got := LowerBound(nil, 5); got != 0 {
		t.Fatalf("LowerBound(nil, 5) = %d, ожидали 0", got)
	}
}

func TestLowerBoundAgainstLinear(t *testing.T) {
	s := make([]int, 0, 200)
	for i := range 200 {
		s = append(s, i/3*2)
	}
	for x := -2; x < 140; x++ {
		want := len(s)
		for i, v := range s {
			if v >= x {
				want = i
				break
			}
		}
		if got := LowerBound(s, x); got != want {
			t.Fatalf("x=%d: получили %d, ожидали %d", x, got, want)
		}
	}
}

func TestLowerBoundIsLogarithmic(t *testing.T) {
	calls := 0
	s := make([]int, 1<<20)
	for i := range s {
		s[i] = i
	}
	start := time.Now()
	for x := range 100000 {
		calls++
		LowerBound(s, x*7)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("%d поисков в слайсе на миллион заняли %v — похоже на линейный поиск", calls, d)
	}
}
