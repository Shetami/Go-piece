package main

func TestListBasic(t *testing.T) {
	var l List[string]
	if l.Len() != 0 {
		t.Fatalf("нулевой список: Len = %d", l.Len())
	}
	for _, s := range []string{"a", "b", "c"} {
		l.PushBack(s)
	}
	var got []string
	for v := range l.All() {
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) || l.Len() != 3 {
		t.Fatalf("получили %q, Len=%d", got, l.Len())
	}
}

func TestListEmptyRange(t *testing.T) {
	var l List[int]
	for v := range l.All() {
		t.Fatalf("пустой список выдал %v", v)
	}
}

func TestListBreak(t *testing.T) {
	var l List[int]
	for i := range 10 {
		l.PushBack(i)
	}
	var got []int
	for v := range l.All() {
		if v == 3 {
			break
		}
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("с break получили %v", got)
	}
}

func TestListCollect(t *testing.T) {
	var l List[int]
	l.PushBack(3)
	l.PushBack(1)
	l.PushBack(2)
	if got := slices.Sorted(l.All()); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("slices.Sorted(l.All()) = %v", got)
	}
}

func TestListPushBackFast(t *testing.T) {
	var l List[int]
	start := time.Now()
	for i := range 200000 {
		l.PushBack(i)
	}
	if d := time.Since(start); d > time.Second || l.Len() != 200000 {
		t.Fatalf("200 000 PushBack заняли %v (Len=%d) — похоже на O(n) на вставку", d, l.Len())
	}
}
