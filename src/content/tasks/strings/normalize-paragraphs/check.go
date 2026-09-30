package main

func chkNorm(t *testing.T, in, want string) {
	t.Helper()
	if got := NormalizeText(in); got != want {
		t.Fatalf("NormalizeText(%q) = %q, ожидали %q", in, got, want)
	}
}

func TestNormalizeInsideParagraph(t *testing.T) {
	chkNorm(t, "  Привет,\t\tмир  ", "Привет, мир")
	chkNorm(t, "первая строка\nвторая   строка", "первая строка вторая строка")
	chkNorm(t, "цена: 100 ₽", "цена: 100 ₽")
}

func TestNormalizeParagraphs(t *testing.T) {
	chkNorm(t, "раз\n\nдва", "раз\n\nдва")
	chkNorm(t, "раз\n\n\n\n  два  \n\n\nтри", "раз\n\nдва\n\nтри")
}

func TestNormalizeBlankWithSpaces(t *testing.T) {
	// Строка из пробелов и табов — тоже пустая строка, то есть разделитель абзацев.
	chkNorm(t, "раз\n   \t\nдва", "раз\n\nдва")
	chkNorm(t, "раз\n \nдва", "раз\n\nдва")
}

func TestNormalizeCRLF(t *testing.T) {
	chkNorm(t, "раз\r\nещё\r\n\r\nдва\r\n", "раз ещё\n\nдва")
}

func TestNormalizeEdges(t *testing.T) {
	chkNorm(t, "", "")
	chkNorm(t, " \n\n \t \n", "")
	chkNorm(t, "\n\n\nтекст\n\n\n", "текст")
}
