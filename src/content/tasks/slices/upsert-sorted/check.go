package main

func chkRows(ids ...int) []Row {
	out := make([]Row, len(ids))
	for i, id := range ids {
		out[i] = Row{id, fmt.Sprint("old", id)}
	}
	return out
}

func TestUpsertBasic(t *testing.T) {
	rows := chkRows(2, 4, 6)
	batch := []Row{{5, "new5"}, {4, "upd4"}, {1, "new1"}, {7, "new7"}}
	b0 := slices.Clone(batch)
	got := Upsert(rows, batch)
	want := []Row{{1, "new1"}, {2, "old2"}, {4, "upd4"}, {5, "new5"}, {6, "old6"}, {7, "new7"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Upsert = %v\nожидали %v", got, want)
	}
	if !reflect.DeepEqual(batch, b0) {
		t.Fatalf("batch изменился: %v", batch)
	}
}

func TestUpsertLastWins(t *testing.T) {
	batch := []Row{{3, "a"}, {1, "x"}, {3, "b"}, {2, "y"}, {3, "c"}, {1, "z"}}
	got := Upsert(chkRows(1, 5), batch)
	want := []Row{{1, "z"}, {2, "y"}, {3, "c"}, {5, "old5"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Upsert = %v, ожидали %v — из повторов в batch побеждает последний", got, want)
	}
	if got := Upsert(nil, []Row{{2, "a"}, {2, "b"}}); !reflect.DeepEqual(got, []Row{{2, "b"}}) {
		t.Fatalf("Upsert(nil, …) = %v", got)
	}
	if got := Upsert(chkRows(1, 2), nil); !reflect.DeepEqual(got, chkRows(1, 2)) {
		t.Fatalf("пустой batch: %v", got)
	}
}

func TestUpsertInPlace(t *testing.T) {
	rows := make([]Row, 3, 10)
	copy(rows, chkRows(10, 20, 30))
	got := Upsert(rows, []Row{{15, "n"}, {35, "n"}, {20, "u"}})
	if len(got) != 5 || &got[0] != &rows[0] {
		t.Fatalf("вместимости rows хватало — результат должен лежать в том же массиве")
	}
	if ids := []int{got[0].ID, got[1].ID, got[2].ID, got[3].ID, got[4].ID}; !reflect.DeepEqual(ids, []int{10, 15, 20, 30, 35}) || got[2].Val != "u" {
		t.Fatalf("Upsert на месте = %v", got)
	}
}

func TestUpsertFast(t *testing.T) {
	n, m := 200000, 20000
	rows := make([]Row, n)
	for i := range rows {
		rows[i] = Row{ID: i * 10}
	}
	batch := make([]Row, m)
	for i := range batch {
		batch[i] = Row{ID: (i*7919)%m*100 + 5, Val: "n"} // все ID новые и разные
	}
	start := time.Now()
	got := Upsert(rows, batch)
	d := time.Since(start)
	if len(got) != n+m || !slices.IsSortedFunc(got, func(a, b Row) int { return cmp.Compare(a.ID, b.ID) }) {
		t.Fatalf("результат неверный: len=%d", len(got))
	}
	if d > 2*time.Second {
		t.Fatalf("%d новых строк в таблицу на %d заняли %v — вставка по одной со сдвигом хвоста?", m, n, d)
	}
}
