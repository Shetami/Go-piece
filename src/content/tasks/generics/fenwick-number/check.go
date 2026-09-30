package main

type chkCents int64

func chkIn(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: зависло — бесконечный цикл по индексу 0?", what)
	}
}

func TestFenwickIndexZero(t *testing.T) {
	chkIn(t, "Add(0, …)", func() {
		f := NewFenwick[int](5)
		f.Add(0, 7)
		f.Add(4, 1)
		if s := f.Sum(1); s != 7 {
			t.Errorf("Sum(1) = %d, ожидали 7 — элемент 0 потерялся", s)
		}
		if s := f.Sum(5); s != 8 {
			t.Errorf("Sum(5) = %d, ожидали 8 — последний элемент потерялся", s)
		}
		if s := f.Sum(0); s != 0 {
			t.Errorf("Sum(0) — пустой префикс, а получили %d", s)
		}
	})
}

func TestFenwickCustomType(t *testing.T) {
	chkIn(t, "Fenwick[Cents]", func() {
		f := NewFenwick[chkCents](4)
		f.Add(1, 150)
		f.Add(2, 250)
		f.Set(1, 100)
		f.Set(3, -30)
		if s := f.RangeSum(1, 4); s != 320 {
			t.Errorf("RangeSum(1, 4) = %d, ожидали 320 (100 + 250 - 30)", s)
		}
		if s := f.RangeSum(3, 1); s != 0 {
			t.Errorf("RangeSum(3, 1) = %d, ожидали 0", s)
		}
		f.Set(2, 250)
		if s := f.Sum(4); s != 320 {
			t.Errorf("Set тем же значением изменил сумму: %d", s)
		}
	})
}

func TestFenwickFloat(t *testing.T) {
	chkIn(t, "Fenwick[float64]", func() {
		f := NewFenwick[float64](3)
		f.Add(0, 0.5)
		f.Add(2, 1.25)
		f.Set(0, 2)
		if s := f.RangeSum(0, 3); s != 3.25 {
			t.Errorf("RangeSum = %v, ожидали 3.25", s)
		}
	})
}

func TestFenwickRandom(t *testing.T) {
	chkIn(t, "случайные операции", func() {
		r := rand.New(rand.NewPCG(7, 8))
		const n = 97
		f := NewFenwick[int64](n)
		ref := make([]int64, n)
		for range 5000 {
			i := r.IntN(n)
			v := int64(r.IntN(2001) - 1000)
			if r.IntN(2) == 0 {
				f.Add(i, v)
				ref[i] += v
			} else {
				f.Set(i, v)
				ref[i] = v
			}
			l, rr := r.IntN(n+1), r.IntN(n+1)
			var want int64
			for j := l; j < rr; j++ {
				want += ref[j]
			}
			if got := f.RangeSum(l, rr); got != want {
				t.Errorf("RangeSum(%d, %d) = %d, ожидали %d", l, rr, got, want)
				return
			}
		}
	})
}

func TestFenwickFast(t *testing.T) {
	chkIn(t, "200 000 операций", func() {
		const n = 200000
		f := NewFenwick[int](n)
		start := time.Now()
		for i := range n {
			f.Add(i, i%10)
			f.RangeSum(i/2, n-i/3)
		}
		if d := time.Since(start); d > time.Second {
			t.Errorf("200 000 Add и RangeSum заняли %v — не похоже на O(log n)", d)
		}
	})
}
