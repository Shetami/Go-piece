package main

func chkIndent(t *testing.T, prefix string, chunks []string, want string) {
	t.Helper()
	var b strings.Builder
	iw := NewIndentWriter(&b, prefix)
	for _, c := range chunks {
		n, err := iw.Write([]byte(c))
		if err != nil || n != len(c) {
			t.Fatalf("Write(%q) = %d, %v; ожидали %d, nil — n считается по байтам p, без префикса", c, n, err, len(c))
		}
	}
	if b.String() != want {
		t.Fatalf("куски %q с префиксом %q дали\n%q\nожидали\n%q", chunks, prefix, b.String(), want)
	}
}

func TestIndentBasic(t *testing.T) {
	chkIndent(t, "  ", []string{"a\nb\n"}, "  a\n  b\n")
	chkIndent(t, "> ", []string{"одна строка"}, "> одна строка")
	chkIndent(t, "  ", []string{"", "x\n", ""}, "  x\n")
	chkIndent(t, "  ", nil, "")
}

func TestIndentChunks(t *testing.T) {
	chkIndent(t, "| ", []string{"пер", "вая\nвто", "рая", "\n", "третья"}, "| первая\n| вторая\n| третья")
	chkIndent(t, "| ", []string{"a", "\n", "\n", "b\n"}, "| a\n\n| b\n")
}

func TestIndentEmptyLines(t *testing.T) {
	chkIndent(t, "\t", []string{"func f() {\n\n\treturn\n}\n\n"}, "\tfunc f() {\n\n\t\treturn\n\t}\n\n")
}

func TestIndentNested(t *testing.T) {
	var b strings.Builder
	outer := NewIndentWriter(&b, "  ")
	inner := NewIndentWriter(outer, "- ")
	fmt.Fprintf(outer, "errors:\n")
	fmt.Fprintf(inner, "первая\nвторая\n")
	fmt.Fprintf(outer, "end\n")
	want := "  errors:\n  - первая\n  - вторая\n  end\n"
	if b.String() != want {
		t.Fatalf("вложенные writer'ы:\n%q\nожидали\n%q", b.String(), want)
	}
}

type chkLimitWriter struct {
	left int
}

func (w *chkLimitWriter) Write(p []byte) (int, error) {
	if len(p) <= w.left {
		w.left -= len(p)
		return len(p), nil
	}
	n := w.left
	w.left = 0
	return n, errors.New("диск заполнен")
}

func TestIndentError(t *testing.T) {
	iw := NewIndentWriter(&chkLimitWriter{left: 10}, ">>")
	p := []byte("aaaa\nbbbb\ncccc\n")
	n, err := iw.Write(p)
	if err == nil || n >= len(p) {
		t.Fatalf("Write при ошибке = %d, %v; ожидали ошибку и n < %d", n, err, len(p))
	}
	if n < 0 || n > 6 {
		t.Fatalf("Write вернул n = %d: в writer ушло 10 байт, из них 4 — префиксы, значит из p не больше 6", n)
	}
	var sb strings.Builder
	if _, err := io.Copy(NewIndentWriter(&sb, "  "), strings.NewReader("x\ny\n")); err != nil {
		t.Fatalf("io.Copy через IndentWriter: %v — n должен равняться len(p), иначе будет short write", err)
	}
}
