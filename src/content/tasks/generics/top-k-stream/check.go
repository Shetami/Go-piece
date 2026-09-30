package main

type chkHit struct {
	URL  string
	Hits int
}

func chkByHits(a, b chkHit) bool { return a.Hits < b.Hits }

func TestTopKBasic(t *testing.T) {
	got := TopK(slices.Values([]int{5, 1, 9, 3, 7, 9, 2}), 3, func(a, b int) bool { return a < b })
	if !reflect.DeepEqual(got, []int{9, 9, 7}) {
		t.Fatalf("TopK(3) = %v, ожидали [9 9 7]", got)
	}
	if got := TopK(slices.Values([]int{2, 1}), 5, func(a, b int) bool { return a < b }); !reflect.DeepEqual(got, []int{2, 1}) {
		t.Fatalf("k больше числа элементов: %v, ожидали [2 1]", got)
	}
	if got := TopK(slices.Values([]int{2, 1}), 0, func(a, b int) bool { return a < b }); got != nil {
		t.Fatalf("k = 0: %v, ожидали nil", got)
	}
	if got := TopK(slices.Values([]string{"b", "a", "c"}), 2, func(a, b string) bool { return a > b }); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("с обратным less — k наименьших: %v, ожидали [a b]", got)
	}
}

func TestTopKTies(t *testing.T) {
	hits := []chkHit{{"/a", 10}, {"/b", 30}, {"/c", 10}, {"/d", 30}, {"/e", 10}, {"/f", 5}, {"/g", 30}, {"/h", 10}}
	got := TopK(slices.Values(hits), 5, chkByHits)
	want := []chkHit{{"/b", 30}, {"/d", 30}, {"/g", 30}, {"/a", 10}, {"/c", 10}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TopK с ничьими = %v, ожидали %v (из равных — встреченные раньше)", got, want)
	}
}

func TestTopKRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(21, 22))
	for range 200 {
		n, k := r.IntN(60), r.IntN(12)
		vals := make([]chkHit, n)
		for i := range vals {
			vals[i] = chkHit{strconv.Itoa(i), r.IntN(8)}
		}
		ref := slices.Clone(vals)
		slices.SortStableFunc(ref, func(a, b chkHit) int { return b.Hits - a.Hits })
		want := ref[:min(k, n)]
		got := TopK(slices.Values(vals), k, chkByHits)
		if len(want) == 0 && len(got) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("n=%d k=%d: TopK = %v, ожидали %v", n, k, got, want)
		}
	}
}

func TestTopKMemory(t *testing.T) {
	const n = 1000000
	seq := func(yield func(int) bool) {
		for i := range n {
			if !yield((i * 7919) % n) {
				return
			}
		}
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	got := TopK(seq, 10, func(a, b int) bool { return a < b })
	runtime.ReadMemStats(&after)
	if len(got) != 10 || got[0] != n-1 || got[9] != n-10 {
		t.Fatalf("TopK(10) из миллиона = %v", got)
	}
	if alloc := after.TotalAlloc - before.TotalAlloc; alloc > 1<<20 {
		t.Fatalf("для k=10 выделено %d байт — поток сохранён целиком, а нужна память O(k)", alloc)
	}
}
