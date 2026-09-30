package main

func chkNaturals(pulled *int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; ; i++ {
			*pulled++
			if !yield(i) {
				return
			}
		}
	}
}

func chkCollect[T any](t *testing.T, seq iter.Seq[T], limit int) (out []T) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("паника при обходе: %v — итератор звал yield после false?", r)
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

func TestSeqChainLazy(t *testing.T) {
	pulled, mapped := 0, 0
	seq := Take(Map(Filter(chkNaturals(&pulled), func(x int) bool { return x%2 == 0 }),
		func(x int) string { mapped++; return strconv.Itoa(x * x) }), 3)
	if pulled != 0 || mapped != 0 {
		t.Fatalf("цепочку только построили, а источник уже тронут: pulled=%d, mapped=%d", pulled, mapped)
	}
	got := chkCollect(t, seq, 100)
	if !reflect.DeepEqual(got, []string{"0", "4", "16"}) {
		t.Fatalf("Take(Map(Filter(0,1,2,...)), 3) = %q, ожидали [0 4 16]", got)
	}
	if pulled != 5 {
		t.Fatalf("из источника взяли %d элементов, ожидали 5 (0..4): после третьего не нужно просить следующий", pulled)
	}
	if mapped != 3 {
		t.Fatalf("f в Map вызвана %d раз, ожидали 3", mapped)
	}
}

func TestSeqTakeZero(t *testing.T) {
	pulled := 0
	if got := chkCollect(t, Take(chkNaturals(&pulled), 0), 10); len(got) != 0 || pulled != 0 {
		t.Fatalf("Take(…, 0): получили %v, из источника взяли %d — ожидали пусто и 0", got, pulled)
	}
	if got := chkCollect(t, Take(chkNaturals(&pulled), -1), 10); len(got) != 0 {
		t.Fatalf("Take(…, -1) = %v, ожидали пусто", got)
	}
}

func TestSeqBreak(t *testing.T) {
	pulled := 0
	seq := Map(Filter(chkNaturals(&pulled), func(x int) bool { return x%3 == 0 }), func(x int) int { return -x })
	if got := chkCollect(t, seq, 4); !reflect.DeepEqual(got, []int{0, -3, -6, -9}) {
		t.Fatalf("первые 4 элемента = %v, ожидали [0 -3 -6 -9]", got)
	}
	if pulled != 10 {
		t.Fatalf("после break из источника взято %d, ожидали 10", pulled)
	}
	short := Take(slices.Values([]int{1, 2}), 5)
	if got := chkCollect(t, short, 100); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Take(2 элемента, 5) = %v, ожидали [1 2]", got)
	}
}

func TestSeqReusable(t *testing.T) {
	seq := Take(Filter(slices.Values([]int{1, 2, 3, 4, 5, 6}), func(x int) bool { return x > 1 }), 2)
	first := chkCollect(t, seq, 100)
	second := chkCollect(t, seq, 100)
	if !reflect.DeepEqual(first, []int{2, 3}) || !reflect.DeepEqual(second, []int{2, 3}) {
		t.Fatalf("два обхода одного Take: %v и %v, ожидали [2 3] оба раза", first, second)
	}
}
