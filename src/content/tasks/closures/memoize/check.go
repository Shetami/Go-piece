package main

func TestMemoizeCachesResults(t *testing.T) {
	calls := 0
	sq := Memoize(func(n int) int { calls++; return n * n })
	for _, n := range []int{3, 3, 4, 3, 4} {
		if got := sq(n); got != n*n {
			t.Fatalf("sq(%d) = %d", n, got)
		}
	}
	if calls != 2 {
		t.Fatalf("f вызвана %d раз, ожидали 2 (по разу на 3 и на 4)", calls)
	}
}

func TestMemoizeCachesZeroValues(t *testing.T) {
	calls := 0
	isEmpty := Memoize(func(s string) bool { calls++; return s == "" })
	isEmpty("x")
	isEmpty("x")
	if calls != 1 {
		t.Fatalf("результат false тоже надо запоминать: f вызвана %d раз", calls)
	}
}

func TestMemoizeIndependent(t *testing.T) {
	a, b := 0, 0
	fa := Memoize(func(n int) int { a++; return n })
	fb := Memoize(func(n int) int { b++; return n })
	fa(1)
	fb(1)
	if a != 1 || b != 1 {
		t.Fatalf("у двух мемоизированных функций должен быть свой кэш: a=%d b=%d", a, b)
	}
}
