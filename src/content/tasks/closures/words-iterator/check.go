package main

func collect(next func() (string, bool)) []string {
	var out []string
	for {
		w, ok := next()
		if !ok {
			return out
		}
		out = append(out, w)
	}
}

func TestWordsBasic(t *testing.T) {
	got := collect(Words("  привет,\tмир \n  go "))
	want := []string{"привет,", "мир", "go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("получили %q, ожидали %q", got, want)
	}
}

func TestWordsEmpty(t *testing.T) {
	for _, s := range []string{"", "   \t\n"} {
		if got := collect(Words(s)); len(got) != 0 {
			t.Fatalf("Words(%q): %q, ожидали ни одного слова", s, got)
		}
	}
}

func TestWordsAfterEnd(t *testing.T) {
	next := Words("a")
	next()
	for range 3 {
		if w, ok := next(); ok || w != "" {
			t.Fatalf("после конца ожидали \"\", false; получили %q, %v", w, ok)
		}
	}
}

func TestWordsIndependent(t *testing.T) {
	a, b := Words("x y"), Words("x y")
	a()
	if w, _ := b(); w != "x" {
		t.Fatalf("итераторы должны быть независимы: второй начал с %q", w)
	}
}

func TestWordsLazy(t *testing.T) {
	big := "первое " + strings.Repeat("слово ", 1<<18)
	allocs := testing.AllocsPerRun(10, func() {
		next := Words(big)
		next()
	})
	if allocs > 3 {
		t.Fatalf("на первое слово ушло %.0f выделений памяти — похоже, строку режут целиком заранее", allocs)
	}
}
