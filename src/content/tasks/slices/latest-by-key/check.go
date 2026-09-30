package main

type chkUpd struct {
	ID  string
	Ver int
}

func TestLatestBasic(t *testing.T) {
	in := []chkUpd{{"a", 1}, {"b", 1}, {"a", 2}, {"c", 1}, {"b", 2}, {"a", 3}}
	orig := slices.Clone(in)
	got := Latest(in, func(u chkUpd) string { return u.ID })
	want := []chkUpd{{"a", 3}, {"b", 2}, {"c", 1}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Latest = %v, ожидали %v: место — по первому вхождению, значение — по последнему", got, want)
	}
	if !reflect.DeepEqual(in, orig) {
		t.Fatalf("вход изменился: %v, был %v", in, orig)
	}
}

func TestLatestNoAlias(t *testing.T) {
	in := []chkUpd{{"a", 1}, {"b", 1}, {"c", 1}}
	got := Latest(in, func(u chkUpd) string { return u.ID })
	got[0].Ver = 100
	_ = append(got[:1], chkUpd{"z", 9})
	if in[0].Ver != 1 || in[1].ID != "b" {
		t.Fatalf("изменение результата попало во вход: %v", in)
	}
}

func TestLatestKeyOnce(t *testing.T) {
	calls := 0
	in := []int{5, 1, 5, 2, 1, 5}
	got := Latest(in, func(v int) int { calls++; return v % 4 })
	if calls != len(in) {
		t.Fatalf("key вызвана %d раз, ожидали %d — ровно по разу на элемент", calls, len(in))
	}
	if want := []int{5, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Latest(%v, v%%4) = %v, ожидали %v", in, got, want)
	}
	if got := Latest([]int(nil), func(v int) int { return v }); len(got) != 0 {
		t.Fatalf("Latest(nil) = %v", got)
	}
}

func TestLatestLinear(t *testing.T) {
	n := 200000
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}
	start := time.Now()
	got := Latest(in, func(v int) int { return v })
	if len(got) != n {
		t.Fatalf("уникальных ключей %d, в результате %d", n, len(got))
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("%d уникальных ключей обработаны за %v — похоже на поиск по результату, а не O(n)", n, d)
	}
}
