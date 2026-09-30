package main

func TestDiffInts(t *testing.T) {
	want := []int{5, 1, 3, 3, 9, 1}
	have := []int{3, 7, 7, 2, 9}
	w0, h0 := slices.Clone(want), slices.Clone(have)
	add, remove := Diff(want, have, cmp.Compare[int])
	if !reflect.DeepEqual(add, []int{1, 5}) || !reflect.DeepEqual(remove, []int{2, 7}) {
		t.Fatalf("Diff = add %v, remove %v; ожидали add [1 5], remove [2 7] — отсортированы и без повторов", add, remove)
	}
	if !reflect.DeepEqual(want, w0) || !reflect.DeepEqual(have, h0) {
		t.Fatalf("входы изменились: want %v, have %v", want, have)
	}
}

func TestDiffEmpty(t *testing.T) {
	add, remove := Diff([]int{2, 1, 2}, nil, cmp.Compare[int])
	if !reflect.DeepEqual(add, []int{1, 2}) || len(remove) != 0 {
		t.Fatalf("have пуст: add %v, remove %v", add, remove)
	}
	add, remove = Diff(nil, []int{4, 4}, cmp.Compare[int])
	if len(add) != 0 || !reflect.DeepEqual(remove, []int{4}) {
		t.Fatalf("want пуст: add %v, remove %v", add, remove)
	}
	add, remove = Diff([]int{1, 2, 2}, []int{2, 1, 1}, cmp.Compare[int])
	if len(add) != 0 || len(remove) != 0 {
		t.Fatalf("одинаковые множества: add %v, remove %v", add, remove)
	}
}

func TestDiffByCmpNotEquality(t *testing.T) {
	fold := func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) }
	want := []string{"Beta", "alpha", "ALPHA", "gamma", "Delta", "DELTA", "delta"}
	have := []string{"BETA", "Epsilon", "epsilon"}
	add, remove := Diff(want, have, fold)
	if wantAdd := []string{"alpha", "Delta", "gamma"}; !reflect.DeepEqual(add, wantAdd) {
		t.Fatalf("add = %q, ожидали %q — равенство по cmp, из равных — первый во входе", add, wantAdd)
	}
	if wantRm := []string{"Epsilon"}; !reflect.DeepEqual(remove, wantRm) {
		t.Fatalf("remove = %q, ожидали %q", remove, wantRm)
	}
}

func TestDiffStableManyEqual(t *testing.T) {
	type rec struct{ K, Seq int }
	want := make([]rec, 200)
	for i := range want {
		want[i] = rec{(i * 7) % 10, i}
	}
	add, _ := Diff(want, nil, func(a, b rec) int { return cmp.Compare(a.K, b.K) })
	for k, r := range add {
		if r.K != k || r.Seq != slices.IndexFunc(want, func(x rec) bool { return x.K == k }) {
			t.Fatalf("add[%d] = %v — из равных по ключу должен остаться первый во входе", k, r)
		}
	}
}

func TestDiffFast(t *testing.T) {
	n := 100000
	want := make([]int, n)
	have := make([]int, n)
	for i := range n {
		want[i] = (i * 7919) % (2 * n)
		have[i] = (i * 104729) % (2 * n)
	}
	start := time.Now()
	Diff(want, have, cmp.Compare[int])
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("Diff на 2×%d элементов занял %v — поиск перебором?", n, d)
	}
}
