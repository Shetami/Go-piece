package main

type chkItem struct {
	key, pos int
}

func TestSortParallelCorrectAndStable(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for _, n := range []int{0, 1, 2, 15, 16, 17, 100, 1000} {
		in := make([]chkItem, n)
		for i := range in {
			in[i] = chkItem{key: r.IntN(10), pos: i}
		}
		orig := slices.Clone(in)
		got := SortParallel(in, 4, func(a, b chkItem) int { return cmp.Compare(a.key, b.key) })
		if !reflect.DeepEqual(in, orig) {
			t.Fatalf("n=%d: входной слайс изменён — нужно вернуть новый", n)
		}
		want := slices.Clone(orig)
		slices.SortStableFunc(want, func(a, b chkItem) int { return cmp.Compare(a.key, b.key) })
		if len(got) != n {
			t.Fatalf("n=%d: длина результата %d", n, len(got))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("n=%d: на позиции %d %v, ожидали %v (равные ключи должны сохранять исходный порядок)", n, i, got[i], want[i])
			}
		}
	}
}

func TestSortParallelLimit(t *testing.T) {
	var active, peak atomic.Int32
	var calls atomic.Int32
	slow := func(a, b int) int {
		n := active.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		if calls.Add(1)%8 == 0 {
			time.Sleep(200 * time.Microsecond)
		}
		active.Add(-1)
		return a - b
	}
	in := rand.Perm(256)
	got := SortParallel(in, 3, slow)
	if !slices.IsSorted(got) {
		t.Fatalf("результат не отсортирован: %v", got[:10])
	}
	if p := peak.Load(); p > 3 {
		t.Fatalf("одновременно сравнивали %d горутин при maxPar=3", p)
	}
	if p := peak.Load(); p < 2 {
		t.Fatalf("сравнения ни разу не шли параллельно — половины сортируются последовательно")
	}
}

func TestSortParallelMaxParOne(t *testing.T) {
	done := make(chan []int, 1)
	go func() { done <- SortParallel([]int{5, 3, 9, 1, 7, 2, 8, 6, 4, 0, 11, 15, 13, 12, 14, 10, 19, 17, 18, 16}, 0, cmp.Compare[int]) }()
	select {
	case got := <-done:
		if !slices.IsSorted(got) {
			t.Fatalf("maxPar=0: %v не отсортирован", got)
		}
	case <-time.After(time.Second):
		t.Fatal("SortParallel с maxPar=0 завис — рекурсия ждёт слот, который никогда не освободится")
	}
}
