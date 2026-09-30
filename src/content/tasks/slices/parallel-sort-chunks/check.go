package main

type chkKV struct{ K, Seq int }

func chkData(n int, seed uint32) []chkKV {
	out := make([]chkKV, n)
	for i := range out {
		seed = seed*1664525 + 1013904223
		out[i] = chkKV{int(seed>>24) % 10, i} // много равных ключей
	}
	return out
}

func chkByKey(a, b chkKV) int { return cmp.Compare(a.K, b.K) }

func TestParallelSortStable(t *testing.T) {
	for _, c := range []struct{ n, workers int }{{1000, 4}, {1001, 3}, {97, 8}, {5, 16}, {1, 3}, {0, 2}, {500, 1}, {2000, 7}} {
		s := chkData(c.n, uint32(c.n+c.workers))
		want := slices.Clone(s)
		slices.SortStableFunc(want, chkByKey)
		ParallelSortFunc(s, c.workers, chkByKey)
		if !reflect.DeepEqual(s, want) {
			for i := range s {
				if s[i] != want[i] {
					t.Fatalf("n=%d, workers=%d: s[%d] = %v, ожидали %v (при равных ключах важен исходный порядок Seq)", c.n, c.workers, i, s[i], want[i])
				}
			}
		}
	}
}

func TestParallelSortWorkersLimit(t *testing.T) {
	var inflight, peak atomic.Int32
	var calls atomic.Int64
	slow := func(a, b chkKV) int {
		calls.Add(1)
		n := inflight.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		runtime.Gosched()
		inflight.Add(-1)
		return chkByKey(a, b)
	}
	s := chkData(4000, 9)
	ParallelSortFunc(s, 3, slow)
	if p := peak.Load(); p > 3 {
		t.Fatalf("cmp одновременно выполнялась в %d горутинах, лимит 3", p)
	}
	if !slices.IsSortedFunc(s, chkByKey) {
		t.Fatalf("результат не отсортирован")
	}
	before := calls.Load()
	time.Sleep(20 * time.Millisecond)
	if after := calls.Load(); after != before {
		t.Fatalf("после возврата cmp вызвали ещё %d раз — горутины не дождались", after-before)
	}
}

func TestParallelSortPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("workers=0: ожидали панику")
		}
	}()
	ParallelSortFunc([]int{2, 1}, 0, cmp.Compare[int])
}
