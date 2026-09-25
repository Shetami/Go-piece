package main

func sameData(a, b string) bool {
	return len(a) == len(b) && (len(a) == 0 || unsafe.StringData(a) == unsafe.StringData(b))
}

func TestInternSame(t *testing.T) {
	var in Interner
	a := in.Intern(strings.Repeat("go", 3))
	b := in.Intern(strings.Repeat("go", 3))
	if a != "gogogo" || !sameData(a, b) {
		t.Fatalf("равные строки должны возвращаться с одними и теми же байтами")
	}
	if in.Len() != 1 {
		t.Fatalf("Len = %d, ожидали 1", in.Len())
	}
	c := in.Intern("rust")
	if c != "rust" || in.Len() != 2 || sameData(a, c) {
		t.Fatalf("разные строки: %q, Len=%d", c, in.Len())
	}
}

func TestInternDetaches(t *testing.T) {
	var in Interner
	big := strings.Repeat("x", 1<<20) + "ERROR"
	word := big[len(big)-5:]
	got := in.Intern(word)
	if got != "ERROR" {
		t.Fatalf("Intern(%q) = %q", word, got)
	}
	if sameData(got, word) {
		t.Fatalf("сохранена подстрока мегабайтной строки — она держит весь мегабайт в памяти")
	}
}

func TestInternZeroValue(t *testing.T) {
	var in Interner
	if in.Len() != 0 {
		t.Fatalf("нулевой Interner: Len = %d", in.Len())
	}
	if in.Intern("") != "" {
		t.Fatalf("пустая строка")
	}
}

func TestInternNoAllocOnHit(t *testing.T) {
	var in Interner
	key := strings.Repeat("k", 32)
	in.Intern(key)
	if a := testing.AllocsPerRun(100, func() { in.Intern(key) }); a != 0 {
		t.Fatalf("повторный Intern выделяет память: %.0f раз", a)
	}
}
