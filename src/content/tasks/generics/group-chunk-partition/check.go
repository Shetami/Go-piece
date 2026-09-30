package main

type chkOrder struct {
	City string
	Sum  int
}

func TestGroupByOrder(t *testing.T) {
	orders := []chkOrder{{"Пермь", 1}, {"Казань", 2}, {"Пермь", 3}, {"Омск", 4}, {"Казань", 5}}
	for range 20 {
		keys, groups := GroupBy(orders, func(o chkOrder) string { return o.City })
		if !reflect.DeepEqual(keys, []string{"Пермь", "Казань", "Омск"}) {
			t.Fatalf("ключи = %q, ожидали в порядке первого появления [Пермь Казань Омск]", keys)
		}
		if !reflect.DeepEqual(groups["Казань"], []chkOrder{{"Казань", 2}, {"Казань", 5}}) || len(groups) != 3 {
			t.Fatalf("groups = %v", groups)
		}
	}
	keys, groups := GroupBy([]int(nil), func(x int) int { return x })
	if len(keys) != 0 || groups == nil {
		t.Fatalf("для пустого входа: keys=%v, groups=%v — мапа должна быть не nil", keys, groups)
	}
	groups[1] = []int{1} // не должно паниковать
}

func TestChunkSizes(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4, 5, 6, 7}, 3)
	if !reflect.DeepEqual(got, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}) {
		t.Fatalf("Chunk(1..7, 3) = %v", got)
	}
	if got := Chunk([]int{1, 2}, 5); !reflect.DeepEqual(got, [][]int{{1, 2}}) {
		t.Fatalf("Chunk(1..2, 5) = %v, ожидали [[1 2]]", got)
	}
	if got := Chunk([]int{}, 3); got != nil {
		t.Fatalf("Chunk(пусто, 3) = %v, ожидали nil", got)
	}
	if got := Chunk([]int{1, 2}, 0); got != nil {
		t.Fatalf("Chunk(…, 0) = %v, ожидали nil", got)
	}
}

func TestChunkIndependent(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	chunks := Chunk(items, 2)
	chunks[0] = append(chunks[0], "X")
	if chunks[1][0] != "c" || items[2] != "c" {
		t.Fatalf("append к первому куску затёр второй: chunks[1]=%q, items=%q", chunks[1], items)
	}
}

func TestPartitionKeepsInput(t *testing.T) {
	items := []int{5, 2, 8, 1, 9, 4}
	orig := slices.Clone(items)
	even, odd := Partition(items, func(x int) bool { return x%2 == 0 })
	if !reflect.DeepEqual(even, []int{2, 8, 4}) || !reflect.DeepEqual(odd, []int{5, 1, 9}) {
		t.Fatalf("Partition = %v, %v, ожидали [2 8 4], [5 1 9]", even, odd)
	}
	if !reflect.DeepEqual(items, orig) {
		t.Fatalf("Partition испортил вход: %v, был %v", items, orig)
	}
}
