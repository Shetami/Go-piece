package main

func chkFib(calls *int) func(int) uint64 {
	return MemoRec(func(self func(int) uint64, n int) uint64 {
		*calls++
		if *calls > 10000 {
			panic("chk: слишком много вызовов")
		}
		if n < 2 {
			return uint64(n)
		}
		return self(n-1) + self(n-2)
	})
}

func TestMemoRecFib(t *testing.T) {
	calls := 0
	fib := chkFib(&calls)
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("fib(90): f вызвана больше 10000 раз — рекурсивные вызовы идут мимо кэша")
			}
		}()
		if v := fib(90); v != 2880067194370816120 {
			t.Fatalf("fib(90) = %d, ожидали 2880067194370816120", v)
		}
	}()
	if calls != 91 {
		t.Fatalf("f вызвана %d раз, ожидали 91 — по разу на каждое n от 0 до 90", calls)
	}
	fib(90)
	fib(0)
	if calls != 91 {
		t.Fatalf("повторные fib(90) и fib(0) вызвали f ещё %d раз — нулевой результат тоже надо помнить", calls-91)
	}
}

func TestMemoRecIndependent(t *testing.T) {
	a, b := 0, 0
	fa, fb := chkFib(&a), chkFib(&b)
	fa(10)
	fb(10)
	if a != 11 || b != 11 {
		t.Fatalf("у каждого MemoRec свой кэш: вызовов %d и %d, ожидали 11 и 11", a, b)
	}
}

func TestMemoRecGridPaths(t *testing.T) {
	type cell struct{ r, c int }
	calls := 0
	paths := MemoRec(func(self func(cell) int, p cell) int {
		calls++
		if p.r == 0 || p.c == 0 {
			return 1
		}
		return self(cell{p.r - 1, p.c}) + self(cell{p.r, p.c - 1})
	})
	if got := paths(cell{16, 16}); got != 601080390 {
		t.Fatalf("путей в сетке 16×16 = %d, ожидали 601080390", got)
	}
	if calls != 17*17-1 {
		t.Fatalf("f вызвана %d раз, ожидали %d — каждая клетка один раз", calls, 17*17-1)
	}
}
