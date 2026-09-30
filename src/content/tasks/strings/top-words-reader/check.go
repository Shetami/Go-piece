package main

func chkTop(t *testing.T, r io.Reader, text string, k int, want []WordCount) {
	t.Helper()
	got, err := TopWords(r, k)
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("TopWords(%q, %d) = %v, %v; ожидали %v", text, k, got, err, want)
	}
}

func chkTopStr(t *testing.T, text string, k int, want []WordCount) {
	t.Helper()
	chkTop(t, strings.NewReader(text), text, k, want)
}

func TestTopBasic(t *testing.T) {
	text := "Go, go, GO! Rust? rust. Zig"
	chkTopStr(t, text, 2, []WordCount{{"go", 3}, {"rust", 2}})
	chkTopStr(t, text, 10, []WordCount{{"go", 3}, {"rust", 2}, {"zig", 1}})
	chkTopStr(t, "b a c a b", 3, []WordCount{{"a", 2}, {"b", 2}, {"c", 1}})
	chkTopStr(t, "", 3, nil)
	chkTopStr(t, "... !!!", 3, nil)
	chkTopStr(t, "a a", 0, nil)
}

func TestTopApostropheHyphen(t *testing.T) {
	chkTopStr(t, "Don't stop. don’t -то кто-то кто - то 'quoted' rock'n'roll' x--y", 20, []WordCount{
		{"то", 2}, {"don't", 1}, {"don’t", 1}, {"quoted", 1}, {"rock'n'roll", 1},
		{"stop", 1}, {"x", 1}, {"y", 1}, {"кто", 1}, {"кто-то", 1},
	})
}

func TestTopYo(t *testing.T) {
	chkTopStr(t, "Ёж ёж ЕЖ еж, ещё ЕЩЕ", 5, []WordCount{{"еж", 4}, {"еще", 2}})
}

type chkByteReader struct{ s string }

func (r *chkByteReader) Read(p []byte) (int, error) {
	if r.s == "" {
		return 0, io.EOF
	}
	p[0] = r.s[0]
	r.s = r.s[1:]
	return 1, nil
}

type chkErrReader struct{ n int }

func (r *chkErrReader) Read(p []byte) (int, error) {
	if r.n == 0 {
		return 0, errors.New("соединение разорвано")
	}
	r.n--
	return copy(p, "слово "), nil
}

func TestTopStreaming(t *testing.T) {
	text := "привет мир привет ёлка-палка"
	chkTop(t, &chkByteReader{text}, text, 3, []WordCount{{"привет", 2}, {"елка-палка", 1}, {"мир", 1}})
	got, err := TopWords(&chkErrReader{n: 3}, 5)
	if err == nil || !strings.Contains(err.Error(), "соединение разорвано") || got != nil {
		t.Fatalf("ошибка чтения: получили %v, %v; ожидали nil и ошибку", got, err)
	}
}
