package main

func feed(n int) <-chan int {
	in := make(chan int)
	go func() {
		defer close(in)
		for i := range n {
			in <- i
		}
	}()
	return in
}

func TestOrderedMapOrder(t *testing.T) {
	// Чем меньше число, тем дольше считается: без упорядочивания результаты перемешаются.
	slow := func(x int) int {
		time.Sleep(time.Duration(20-x) * time.Millisecond)
		return x * x
	}
	var got []int
	for v := range OrderedMap(feed(20), 4, slow) {
		got = append(got, v)
	}
	for i := range 20 {
		if i >= len(got) || got[i] != i*i {
			t.Fatalf("получили %v, ожидали квадраты 0..19 по порядку", got)
		}
	}
}

func TestOrderedMapParallel(t *testing.T) {
	start := time.Now()
	n := 0
	for range OrderedMap(feed(16), 8, func(x int) int { time.Sleep(50 * time.Millisecond); return x }) {
		n++
	}
	if n != 16 {
		t.Fatalf("получили %d результатов, ожидали 16", n)
	}
	if d := time.Since(start); d > 400*time.Millisecond {
		t.Fatalf("16 задач по 50 мс с 8 воркерами заняли %v — похоже, работают по одной", d)
	}
}

func TestOrderedMapLimit(t *testing.T) {
	var active, peak atomic.Int32
	f := func(x int) int {
		n := active.Add(1)
		for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
		}
		time.Sleep(10 * time.Millisecond)
		active.Add(-1)
		return x
	}
	for range OrderedMap(feed(30), 3, f) {
	}
	if p := peak.Load(); p > 3 {
		t.Fatalf("одновременно работало %d вызовов, лимит 3", p)
	}
}

func TestOrderedMapEmpty(t *testing.T) {
	in := make(chan int)
	close(in)
	select {
	case _, ok := <-OrderedMap(in, 2, func(x int) int { return x }):
		if ok {
			t.Fatalf("из пустого входа пришло значение")
		}
	case <-time.After(time.Second):
		t.Fatalf("выход не закрылся после закрытия пустого входа")
	}
}
