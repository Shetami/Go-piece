package main

func chkSuggest(t *testing.T, in string, cands []string, maxDist, limit int, want []string) {
	t.Helper()
	got := Suggest(in, cands, maxDist, limit)
	if !slices.Equal(got, want) {
		t.Fatalf("Suggest(%q, %q, %d, %d) = %q, ожидали %q", in, cands, maxDist, limit, got, want)
	}
}

func TestSuggestBasic(t *testing.T) {
	cmds := []string{"build", "test", "vet", "run", "fmt", "get", "list"}
	chkSuggest(t, "biuld", cmds, 2, 3, []string{"build"})
	// Перестановка соседних букв — две операции, а не одна.
	chkSuggest(t, "tets", cmds, 2, 3, []string{"get", "test", "vet"})
	chkSuggest(t, "xyzzy", cmds, 2, 3, nil)
	chkSuggest(t, "ru", cmds, 1, 5, []string{"run"})
	chkSuggest(t, "", cmds, 3, 2, []string{"fmt", "get"})
}

func TestSuggestRunes(t *testing.T) {
	// В байтах «ё»→«е» — две замены, в символах — одна.
	words := []string{"ёлка", "полка", "белка", "метла"}
	chkSuggest(t, "елка", words, 1, 5, []string{"белка", "ёлка"})
	chkSuggest(t, "плка", words, 1, 5, []string{"полка", "ёлка"})
}

func TestSuggestCaseAndDup(t *testing.T) {
	cands := []string{"Deploy", "deploy", "DEPLOY", "Delay", "replay"}
	chkSuggest(t, "DEPLOI", cands, 3, 5, []string{"Deploy", "Delay", "replay"})
}

func TestSuggestTieOrderAndLimit(t *testing.T) {
	cands := []string{"cat", "bat", "hat", "at", "mat", "chat"}
	chkSuggest(t, "xat", cands, 1, 10, []string{"at", "bat", "cat", "hat", "mat"})
	chkSuggest(t, "xat", cands, 1, 2, []string{"at", "bat"})
	chkSuggest(t, "xat", cands, 1, 0, nil)
}
