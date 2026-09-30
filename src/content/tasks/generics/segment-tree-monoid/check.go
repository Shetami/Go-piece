package main

func chkConcat(a, b string) string { return a + b }

func TestSegMin(t *testing.T) {
	vals := []int{5, 3, 8, 6, 1, 9, 2}
	s := NewSegTree(vals, math.MaxInt, func(a, b int) int { return min(a, b) })
	if q := s.Query(0, 7); q != 1 {
		t.Fatalf("Query(0, 7) = %d, ожидали 1", q)
	}
	if q := s.Query(0, 4); q != 3 {
		t.Fatalf("Query(0, 4) = %d, ожидали 3", q)
	}
	if q := s.Query(4, 4); q != math.MaxInt {
		t.Fatalf("пустой отрезок — нейтральный элемент, а получили %d", q)
	}
	vals[1] = -100
	if q := s.Query(0, 3); q != 3 {
		t.Fatalf("после правки исходного слайса Query(0, 3) = %d — дерево делит с ним память", q)
	}
	s.Set(6, 0)
	if q := s.Query(5, 7); q != 0 {
		t.Fatalf("после Set(6, 0) Query(5, 7) = %d", q)
	}
}

func TestSegOrderMatters(t *testing.T) {
	letters := strings.Split("abcdefghijk", "")
	s := NewSegTree(letters, "", chkConcat)
	for _, c := range [][2]int{{0, 11}, {1, 10}, {3, 8}, {2, 3}, {5, 11}} {
		want := strings.Join(letters[c[0]:c[1]], "")
		if got := s.Query(c[0], c[1]); got != want {
			t.Fatalf("Query(%d, %d) = %q, ожидали %q — порядок склейки нарушен", c[0], c[1], got, want)
		}
	}
}

func TestSegRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(11, 12))
	for _, n := range []int{1, 2, 13, 64, 100} {
		ref := make([]string, n)
		for i := range ref {
			ref[i] = string(rune('a' + r.IntN(26)))
		}
		s := NewSegTree(slices.Clone(ref), "", chkConcat)
		for range 500 {
			if r.IntN(3) == 0 {
				i := r.IntN(n)
				ref[i] = string(rune('A' + r.IntN(26)))
				s.Set(i, ref[i])
			}
			l := r.IntN(n + 1)
			rr := l + r.IntN(n-l+1)
			if got, want := s.Query(l, rr), strings.Join(ref[l:rr], ""); got != want {
				t.Fatalf("n=%d: Query(%d, %d) = %q, ожидали %q", n, l, rr, got, want)
			}
		}
	}
}

func TestSegFast(t *testing.T) {
	const n = 100000
	vals := make([]int, n)
	for i := range vals {
		vals[i] = i
	}
	start := time.Now()
	s := NewSegTree(vals, 0, func(a, b int) int { return a + b })
	sum := 0
	for i := range n {
		s.Set(i, 1)
		sum += s.Query(i/2, n-i/2)
	}
	if sum == 0 {
		t.Fatalf("сумма не посчиталась")
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("100 000 Set и Query заняли %v — не похоже на O(log n)", d)
	}
}
