package main

func chkAdd(s *Set, iv ...[2]int) {
	for _, p := range iv {
		s.Add(p[0], p[1])
	}
}

func TestSetMerge(t *testing.T) {
	cases := []struct {
		add  [][2]int
		want []Interval
	}{
		{[][2]int{{1, 3}, {5, 7}}, []Interval{{1, 3}, {5, 7}}},
		{[][2]int{{5, 7}, {1, 3}}, []Interval{{1, 3}, {5, 7}}},
		{[][2]int{{1, 3}, {3, 5}}, []Interval{{1, 5}}},                  // касание
		{[][2]int{{3, 5}, {1, 3}}, []Interval{{1, 5}}},                  // касание слева
		{[][2]int{{1, 2}, {4, 5}, {7, 8}, {10, 11}, {3, 9}}, []Interval{{1, 2}, {3, 9}, {10, 11}}},
		{[][2]int{{1, 2}, {4, 5}, {7, 8}, {2, 7}}, []Interval{{1, 8}}},  // мост через несколько
		{[][2]int{{1, 10}, {3, 4}}, []Interval{{1, 10}}},                // внутри — ничего не меняет
		{[][2]int{{1, 3}, {5, 5}, {7, 6}}, []Interval{{1, 3}}},          // пустые игнорируются
		{[][2]int{{5, 6}, {0, 100}}, []Interval{{0, 100}}},
	}
	for _, c := range cases {
		var s Set
		chkAdd(&s, c.add...)
		if got := s.Intervals(); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("после Add%v интервалы %v, ожидали %v", c.add, got, c.want)
		}
	}
}

func TestSetContains(t *testing.T) {
	var s Set
	chkAdd(&s, [2]int{10, 20}, [2]int{30, 31})
	for x, want := range map[int]bool{9: false, 10: true, 19: true, 20: false, 25: false, 30: true, 31: false, -5: false} {
		if got := s.Contains(x); got != want {
			t.Fatalf("Contains(%d) = %v, ожидали %v; интервалы %v (End не входит)", x, got, want, s.Intervals())
		}
	}
	var empty Set
	if empty.Contains(0) {
		t.Fatalf("пустое множество содержит 0")
	}
}

func TestSetIntervalsCopy(t *testing.T) {
	var s Set
	s.Add(1, 5)
	iv := s.Intervals()
	iv[0].End = 100
	if s.Contains(50) {
		t.Fatalf("изменение результата Intervals попало внутрь множества")
	}
}

func TestSetAgainstBitmap(t *testing.T) {
	var s Set
	var bits [200]bool
	seed := uint32(7)
	for range 400 {
		seed = seed*1664525 + 1013904223
		a := int(seed>>8) % 200
		seed = seed*1664525 + 1013904223
		l := int(seed>>8)%12 - 2
		s.Add(a, a+l)
		for x := a; x < min(a+l, 200); x++ {
			bits[x] = true
		}
		iv := s.Intervals()
		for k := 1; k < len(iv); k++ {
			if iv[k-1].End >= iv[k].Start {
				t.Fatalf("интервалы %v и %v должны были склеиться", iv[k-1], iv[k])
			}
		}
		for x := range 200 {
			if s.Contains(x) != bits[x] {
				t.Fatalf("после Add(%d,%d) Contains(%d) = %v, ожидали %v", a, a+l, x, !bits[x], bits[x])
			}
		}
	}
}

func TestSetFast(t *testing.T) {
	var s Set
	start := time.Now()
	for i := range 20000 {
		s.Add(i*10, i*10+5)
	}
	hits := 0
	for q := range 300000 {
		if s.Contains(q * 7 % 200000) {
			hits++
		}
	}
	if len(s.Intervals()) != 20000 || hits == 0 {
		t.Fatalf("ожидали 20000 интервалов, получили %d", len(s.Intervals()))
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("20000 Add и 300000 Contains заняли %v — поиск не двоичный?", d)
	}
}
