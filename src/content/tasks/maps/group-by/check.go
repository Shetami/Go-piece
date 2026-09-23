package main

func TestGroupByLength(t *testing.T) {
	got := GroupBy([]string{"go", "c", "rust", "zig", "js", "d"}, func(s string) int { return len(s) })
	want := map[int][]string{1: {"c", "d"}, 2: {"go", "js"}, 3: {"zig"}, 4: {"rust"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GroupBy по длине = %v, ожидали %v", got, want)
	}
}

func TestGroupByStructs(t *testing.T) {
	type order struct {
		User string
		Sum  int
	}
	orders := []order{{"аня", 10}, {"боря", 5}, {"аня", 7}}
	got := GroupBy(orders, func(o order) string { return o.User })
	if want := []order{{"аня", 10}, {"аня", 7}}; !reflect.DeepEqual(got["аня"], want) {
		t.Fatalf("группа аня = %v, ожидали %v (порядок как во входе)", got["аня"], want)
	}
	if len(got) != 2 {
		t.Fatalf("ожидали две группы, получили %d", len(got))
	}
}

func TestGroupByEmpty(t *testing.T) {
	got := GroupBy([]int(nil), func(n int) int { return n })
	if got == nil {
		t.Fatal("для пустого входа ожидали пустую мапу, а не nil: в неё должно быть можно писать")
	}
	got[1] = append(got[1], 1)
}
