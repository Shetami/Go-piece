package main

type chkEv struct {
	TS  int
	Tag string
}

func chkByTS(a, b chkEv) int { return cmp.Compare(a.TS, b.TS) }

// chkCounting — бесконечный источник step, 2*step, …; считает выданные
// элементы и завершения.
func chkCounting(step int, pulled, finished *int) iter.Seq[int] {
	return func(yield func(int) bool) {
		defer func() { *finished++ }()
		for v := step; ; v += step {
			*pulled++
			if !yield(v) {
				return
			}
		}
	}
}

func chkFirst[T any](t *testing.T, seq iter.Seq[T], limit int) (out []T) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("паника при обходе: %v", r)
		}
	}()
	for v := range seq {
		out = append(out, v)
		if len(out) == limit {
			break
		}
	}
	return out
}

func TestMergeBasic(t *testing.T) {
	got := chkFirst(t, Merge(cmp.Compare[int],
		slices.Values([]int{1, 4, 9}), slices.Values([]int{}), slices.Values([]int{2, 3, 10, 11}), slices.Values([]int{5})), 100)
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5, 9, 10, 11}) {
		t.Fatalf("Merge = %v", got)
	}
	if got := chkFirst(t, Merge[int](cmp.Compare[int]), 10); len(got) != 0 {
		t.Fatalf("Merge без источников = %v", got)
	}
}

func TestMergeStable(t *testing.T) {
	a := []chkEv{{1, "a1"}, {2, "a2"}, {2, "a3"}}
	b := []chkEv{{1, "b1"}, {2, "b2"}}
	c := []chkEv{{0, "c0"}, {1, "c1"}}
	var tags []string
	for e := range Merge(chkByTS, slices.Values(c), slices.Values(a), slices.Values(b)) {
		tags = append(tags, e.Tag)
	}
	want := []string{"c0", "c1", "a1", "b1", "a2", "a3", "b2"}
	if !reflect.DeepEqual(tags, want) {
		t.Fatalf("порядок равных = %q, ожидали %q (по номеру источника, внутри — как было)", tags, want)
	}
}

func TestMergeLazyAndStops(t *testing.T) {
	var pulled, finished int
	m := Merge(cmp.Compare[int], chkCounting(3, &pulled, &finished), chkCounting(5, &pulled, &finished), chkCounting(7, &pulled, &finished))
	if pulled != 0 {
		t.Fatalf("Merge только построили, а из источников уже взято %d", pulled)
	}
	got := chkFirst(t, m, 6)
	if !reflect.DeepEqual(got, []int{3, 5, 6, 7, 9, 10}) {
		t.Fatalf("первые 6 из бесконечных источников = %v", got)
	}
	if pulled > 6+3 {
		t.Fatalf("для 6 элементов из источников взято %d — больше одного впрок на источник", pulled)
	}
	if finished != 3 {
		t.Fatalf("после break остановлено %d источников из 3 — не вызван stop у iter.Pull?", finished)
	}
}

func TestMergeManySources(t *testing.T) {
	const k, n = 5000, 40
	seqs := make([]iter.Seq[int], k)
	for i := range k {
		seqs[i] = func(yield func(int) bool) {
			for j := range n {
				if !yield(j*k + (i*7919)%k) {
					return
				}
			}
		}
	}
	start := time.Now()
	prev, cnt := -1, 0
	for v := range Merge(cmp.Compare[int], seqs...) {
		if v < prev {
			t.Fatalf("после %d идёт %d — порядок нарушен", prev, v)
		}
		prev = v
		cnt++
	}
	if cnt != k*n {
		t.Fatalf("выдано %d элементов, ожидали %d", cnt, k*n)
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("слияние %d источников заняло %v — минимум ищется перебором, нужна куча", k, d)
	}
}
