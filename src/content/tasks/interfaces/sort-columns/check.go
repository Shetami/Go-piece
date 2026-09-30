package main

func chkTable() *Table {
	return &Table{
		Names:  []string{"dan", "ann", "bob", "cat", "eve", "fox"},
		Ages:   []int{30, 25, 30, 25, 30, 41},
		Scores: []float64{7, 9, math.NaN(), 9, 7, 5},
	}
}

func chkRows(t *Table) string {
	var b strings.Builder
	for i := range t.Names {
		fmt.Fprintf(&b, "%s/%d/%v ", t.Names[i], t.Ages[i], t.Scores[i])
	}
	return strings.TrimSpace(b.String())
}

func TestSortSwapsRows(t *testing.T) {
	tb := chkTable()
	if err := SortBy(tb, "name"); err != nil {
		t.Fatal(err)
	}
	want := "ann/25/9 bob/30/NaN cat/25/9 dan/30/7 eve/30/7 fox/41/5"
	if got := chkRows(tb); got != want {
		t.Fatalf("SortBy(name):\n  получили %s\n  ожидали  %s\n(строка переставляется целиком, со всеми колонками)", got, want)
	}
}

func TestSortMultiStable(t *testing.T) {
	tb := chkTable()
	SortBy(tb, "-age", "score")
	want := "fox/41/5 dan/30/7 eve/30/7 bob/30/NaN ann/25/9 cat/25/9"
	if got := chkRows(tb); got != want {
		t.Fatalf("SortBy(-age, score):\n  получили %s\n  ожидали  %s", got, want)
	}
	tb = chkTable()
	SortBy(tb, "-score")
	want = "ann/25/9 cat/25/9 dan/30/7 eve/30/7 fox/41/5 bob/30/NaN"
	if got := chkRows(tb); got != want {
		t.Fatalf("SortBy(-score): равные должны сохранить исходный порядок, NaN — в конце:\n  получили %s\n  ожидали  %s", got, want)
	}
}

func TestSortNaNAscending(t *testing.T) {
	tb := &Table{Names: []string{"a", "b", "c", "d"}, Ages: make([]int, 4), Scores: []float64{math.NaN(), 2, math.NaN(), 1}}
	SortBy(tb, "score")
	if got := chkRows(tb); got != "d/0/1 b/0/2 a/0/NaN c/0/NaN" {
		t.Fatalf("SortBy(score) с NaN: %s; ожидали NaN в конце в исходном порядке", got)
	}
}

func TestSortBadKeys(t *testing.T) {
	tb := chkTable()
	before := chkRows(tb)
	if err := SortBy(tb, "age", "salary"); err == nil {
		t.Fatal("неизвестный ключ salary должен давать ошибку")
	}
	if err := SortBy(tb); err == nil {
		t.Fatal("пустой список ключей должен давать ошибку")
	}
	if chkRows(tb) != before {
		t.Fatalf("после ошибки таблица изменилась: %s", chkRows(tb))
	}
	rag := &Table{Names: []string{"a", "b"}, Ages: []int{1}, Scores: []float64{1, 2}}
	if err := SortBy(rag, "name"); !errors.Is(err, ErrRagged) {
		t.Fatalf("колонки разной длины: %v, ожидали ErrRagged", err)
	}
}
