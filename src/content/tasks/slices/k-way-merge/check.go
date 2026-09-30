package main

type chkRec struct{ Key, Src int }

func TestMergeKInts(t *testing.T) {
	lists := [][]int{{1, 4, 9}, nil, {2, 3, 10, 11}, {}, {0, 5}}
	orig := [][]int{{1, 4, 9}, nil, {2, 3, 10, 11}, {}, {0, 5}}
	got := MergeK(lists, cmp.Compare[int])
	if want := []int{0, 1, 2, 3, 4, 5, 9, 10, 11}; !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeK = %v, ожидали %v", got, want)
	}
	if !reflect.DeepEqual(lists, orig) {
		t.Fatalf("списки изменились: %v", lists)
	}
	if got := MergeK([][]int{nil, {}}, cmp.Compare[int]); len(got) != 0 {
		t.Fatalf("MergeK(пустые) = %v", got)
	}
	if got := MergeK([][]int{{3, 3, 3}}, cmp.Compare[int]); !reflect.DeepEqual(got, []int{3, 3, 3}) {
		t.Fatalf("один список: %v", got)
	}
}

func TestMergeKStable(t *testing.T) {
	lists := [][]chkRec{
		{{1, 10}, {2, 10}, {2, 11}},
		{{1, 20}, {2, 20}},
		{{2, 30}, {3, 30}},
	}
	got := MergeK(lists, func(a, b chkRec) int { return cmp.Compare(a.Key, b.Key) })
	want := []chkRec{{1, 10}, {1, 20}, {2, 10}, {2, 11}, {2, 20}, {2, 30}, {3, 30}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeK = %v\nожидали %v\n(равные ключи — по номеру списка, внутри списка — по порядку)", got, want)
	}
}

func chkCountedMerge(t *testing.T, k, per int) (int, int) {
	t.Helper()
	lists := make([][]int, k)
	for i := range lists {
		lists[i] = make([]int, per)
		for j := range lists[i] {
			lists[i][j] = j*k + (i*7919)%k
		}
	}
	calls := 0
	got := MergeK(lists, func(a, b int) int { calls++; return cmp.Compare(a, b) })
	if len(got) != k*per || !slices.IsSorted(got) {
		t.Fatalf("k=%d: результат не отсортирован или неполон", k)
	}
	return calls, k * per
}

func TestMergeKManyLists(t *testing.T) {
	// k=256: log k = 8. Линейный выбор минимума — ~255 сравнений на элемент.
	calls, n := chkCountedMerge(t, 256, 40)
	if calls > n*(3*8+4) {
		t.Fatalf("k=256, N=%d: %d сравнений (%.1f на элемент) — ожидали O(log k) на элемент", n, calls, float64(calls)/float64(n))
	}
}

func TestMergeKFewLongLists(t *testing.T) {
	// k=4: log k = 2. Склеить всё и отсортировать — ~log N ≈ 16 сравнений на элемент.
	calls, n := chkCountedMerge(t, 4, 16000)
	if calls > n*(3*2+4) {
		t.Fatalf("k=4, N=%d: %d сравнений (%.1f на элемент) — похоже на сортировку всего, а не на слияние", n, calls, float64(calls)/float64(n))
	}
}
