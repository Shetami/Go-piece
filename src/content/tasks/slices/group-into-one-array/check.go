package main

type chkOrder struct {
	City string
	N    int
}

func TestGroupByOrder(t *testing.T) {
	in := []chkOrder{{"msk", 1}, {"spb", 2}, {"msk", 3}, {"kzn", 4}, {"spb", 5}, {"msk", 6}}
	orig := slices.Clone(in)
	calls := 0
	got := GroupBy(in, func(o chkOrder) string { calls++; return o.City })
	want := []Group[string, chkOrder]{
		{"msk", []chkOrder{{"msk", 1}, {"msk", 3}, {"msk", 6}}},
		{"spb", []chkOrder{{"spb", 2}, {"spb", 5}}},
		{"kzn", []chkOrder{{"kzn", 4}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GroupBy = %v\nожидали %v (группы — по первому появлению ключа)", got, want)
	}
	if calls != len(in) {
		t.Fatalf("key вызвана %d раз, ожидали %d", calls, len(in))
	}
	if !reflect.DeepEqual(in, orig) {
		t.Fatalf("вход изменился: %v", in)
	}
	if got := GroupBy([]int(nil), func(v int) int { return v }); len(got) != 0 {
		t.Fatalf("GroupBy(nil) = %v", got)
	}
}

func TestGroupByIndependent(t *testing.T) {
	got := GroupBy([]int{1, 2, 3, 4, 5, 6}, func(v int) int { return v % 3 })
	// группы: 1 → [1 4], 2 → [2 5], 0 → [3 6]
	_ = append(got[0].Items, 100)
	if !reflect.DeepEqual(got[1].Items, []int{2, 5}) {
		t.Fatalf("append к первой группе затёр вторую: %v", got[1].Items)
	}
}

func TestGroupBySharedBacking(t *testing.T) {
	got := GroupBy([]int{1, 2, 3, 4, 5, 6, 7}, func(v int) bool { return v%2 == 0 })
	if len(got) != 2 {
		t.Fatalf("ожидали 2 группы, получили %d", len(got))
	}
	a, b := got[0].Items, got[1].Items
	end := uintptr(unsafe.Pointer(&a[0])) + uintptr(len(a))*unsafe.Sizeof(a[0])
	if uintptr(unsafe.Pointer(&b[0])) != end {
		t.Fatalf("группы должны лежать подряд в одном общем массиве")
	}
}

func TestGroupByAllocs(t *testing.T) {
	in := make([]int, 3000)
	for i := range in {
		in[i] = i
	}
	allocs := testing.AllocsPerRun(10, func() {
		GroupBy(in, func(v int) int { return v % 4 })
	})
	if allocs > 16 {
		t.Fatalf("3000 элементов в 4 группы — %.0f аллокаций; слайсы групп растут через append по одному?", allocs)
	}
}
