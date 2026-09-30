package main

func chkTable(t *testing.T, rows [][]string, want string) {
	t.Helper()
	var b strings.Builder
	if err := WriteTable(&b, rows); err != nil {
		t.Fatalf("WriteTable: неожиданная ошибка %v", err)
	}
	if b.String() != want {
		t.Fatalf("WriteTable(%q):\n%s\nожидали:\n%s", rows, b.String(), want)
	}
}

func TestTableCyrillic(t *testing.T) {
	chkTable(t, [][]string{
		{"товар", "цена", "склад"},
		{"чай", "90", "Москва"},
		{"coffee", "1350.50", "Kazan"},
		{"сахар", "-5", "—"},
	}, ""+
		"товар   цена     склад\n"+
		"------  -------  ------\n"+
		"чай          90  Москва\n"+
		"coffee  1350.50  Kazan\n"+
		"сахар        -5  —\n")
}

func TestTableNumbersAndText(t *testing.T) {
	// Заголовок "id" не число; "v1.2", "1.", ".5", "1e3" — не числа.
	chkTable(t, [][]string{
		{"id", "ver"},
		{"7", "v1.2"},
		{"12", "1."},
		{"3", ".5"},
		{"100", "1e3"},
	}, ""+
		"id   ver\n"+
		"---  ----\n"+
		"  7  v1.2\n"+
		" 12  1.\n"+
		"  3  .5\n"+
		"100  1e3\n")
}

func TestTableRagged(t *testing.T) {
	chkTable(t, [][]string{
		{"a", "b"},
		{"длинное"},
		{"x", "y", "z"},
	}, ""+
		"a        b\n"+
		"-------  -  -\n"+
		"длинное\n"+
		"x        y  z\n")
	var b strings.Builder
	if err := WriteTable(&b, nil); err != nil || b.Len() != 0 {
		t.Fatalf("пустая таблица: %q, %v", b.String(), err)
	}
}

type chkFailWriter struct{}

func (chkFailWriter) Write(p []byte) (int, error) { return 0, errors.New("диск заполнен") }

func TestTableWriterError(t *testing.T) {
	err := WriteTable(chkFailWriter{}, [][]string{{"a"}, {"1"}})
	if err == nil || !strings.Contains(err.Error(), "диск заполнен") {
		t.Fatalf("ошибка записи потерялась: %v", err)
	}
}
