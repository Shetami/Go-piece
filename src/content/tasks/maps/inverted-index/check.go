package main

func chkIDs(t *testing.T, q string, got, want []int) {
	t.Helper()
	if len(got) != len(want) || (len(got) > 0 && !reflect.DeepEqual(got, want)) {
		t.Fatalf("Search(%q) = %v, ожидали %v", q, got, want)
	}
}

func newChkIndex() *Index {
	ix := NewIndex()
	ix.Add(3, "Go, горутины и каналы")
	ix.Add(1, "Каналы в GO: буферизованные и нет")
	ix.Add(2, "Мапы в Go")
	ix.Add(7, "SQL и индексы")
	return ix
}

func TestIndexAnd(t *testing.T) {
	ix := newChkIndex()
	chkIDs(t, "go", ix.Search("go"), []int{1, 2, 3})
	chkIDs(t, "КАНАЛЫ go", ix.Search("КАНАЛЫ go"), []int{1, 3})
	chkIDs(t, "go каналы go", ix.Search("go каналы go"), []int{1, 3})
	chkIDs(t, "go rust", ix.Search("go rust"), nil)
	chkIDs(t, "  ,;! ", ix.Search("  ,;! "), nil)
	chkIDs(t, "", ix.Search(""), nil)
}

func TestIndexReAdd(t *testing.T) {
	ix := newChkIndex()
	ix.Add(2, "Слайсы в Rust")
	chkIDs(t, "мапы", ix.Search("мапы"), nil)
	chkIDs(t, "go", ix.Search("go"), []int{1, 3})
	chkIDs(t, "слайсы", ix.Search("слайсы"), []int{2})
	ix.Add(2, "Слайсы в Rust") // тот же текст ещё раз — ничего не ломается
	chkIDs(t, "rust слайсы", ix.Search("rust слайсы"), []int{2})
}

func TestIndexRemoveCleansWords(t *testing.T) {
	ix := NewIndex()
	ix.Add(1, "alpha beta beta")
	ix.Add(2, "beta gamma")
	if ix.Words() != 3 {
		t.Fatalf("Words = %d, ожидали 3 (alpha, beta, gamma)", ix.Words())
	}
	ix.Remove(1)
	ix.Remove(1)
	ix.Remove(42)
	if ix.Words() != 2 {
		t.Fatalf("после Remove(1): Words = %d, ожидали 2 — слово alpha больше нигде не встречается", ix.Words())
	}
	chkIDs(t, "beta", ix.Search("beta"), []int{2})
	ix.Add(2, "")
	if ix.Words() != 0 {
		t.Fatalf("после замены текста на пустой Words = %d, ожидали 0", ix.Words())
	}
}

func TestIndexResultIsCopy(t *testing.T) {
	ix := newChkIndex()
	got := ix.Search("go")
	for i := range got {
		got[i] = -1
	}
	chkIDs(t, "go", ix.Search("go"), []int{1, 2, 3})
}

func TestIndexMany(t *testing.T) {
	ix := NewIndex()
	for i := range 300 {
		ix.Add(i, fmt.Sprintf("doc a%d b%d", i%3, i%5))
	}
	got := ix.Search("a0 b0 A0 doc")
	for _, id := range got {
		if id%3 != 0 || id%5 != 0 {
			t.Fatalf("документ %d не содержит a0 и b0", id)
		}
	}
	if len(got) != 20 || !slices.IsSorted(got) {
		t.Fatalf("Search = %d документов (отсортирован: %v), ожидали 20 по возрастанию", len(got), slices.IsSorted(got))
	}
}
