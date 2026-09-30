package main

func chkHL(t *testing.T, text string, terms []string, want string) {
	t.Helper()
	if got := Highlight(text, terms); got != want {
		t.Fatalf("Highlight(%q, %q) =\n  %q\nожидали\n  %q", text, terms, got, want)
	}
}

func TestHighlightBasic(t *testing.T) {
	chkHL(t, "Язык Go и горутины", []string{"GO", "го"}, "Язык <mark>Go</mark> и <mark>го</mark>рутины")
	chkHL(t, "ГО, гО, го", []string{"Го"}, "<mark>ГО</mark>, <mark>гО</mark>, <mark>го</mark>")
	chkHL(t, "ничего нет", []string{"xyz", ""}, "ничего нет")
	chkHL(t, "текст", nil, "текст")
}

func TestHighlightMerge(t *testing.T) {
	chkHL(t, "abcde", []string{"abc", "cde"}, "<mark>abcde</mark>")
	chkHL(t, "abcd", []string{"ab", "cd"}, "<mark>abcd</mark>")
	chkHL(t, "ab-cd", []string{"cd", "ab"}, "<mark>ab</mark>-<mark>cd</mark>")
	chkHL(t, "ааа б", []string{"аа"}, "<mark>ааа</mark> б")
	chkHL(t, "strings.Builder", []string{"string", "strings", "build"}, "<mark>strings</mark>.<mark>Build</mark>er")
}

func TestHighlightLengthChangingLower(t *testing.T) {
	// Знак кельвина U+212A (3 байта) в нижнем регистре — 'k' (1 байт),
	// 'İ' (2 байта) — "i̇" (3 байта). Индексы в strings.ToLower(text) съезжают.
	chkHL(t, "5K и Go", []string{"go"}, "5K и <mark>Go</mark>")
	chkHL(t, "İİİ go go", []string{"go"}, "İİİ <mark>go</mark> <mark>go</mark>")
}

func TestHighlightEscape(t *testing.T) {
	chkHL(t, `<b>a & "b"</b>`, []string{"a & \"b"},
		`&lt;b&gt;<mark>a &amp; &quot;b</mark>&quot;&lt;/b&gt;`)
	chkHL(t, "x<y", []string{"lt"}, "x&lt;y")
}
