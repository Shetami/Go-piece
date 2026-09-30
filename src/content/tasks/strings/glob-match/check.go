package main

func chkGlob(t *testing.T, pattern, name string, want bool) {
	t.Helper()
	got, err := Match(pattern, name)
	if err != nil || got != want {
		t.Fatalf("Match(%q, %q) = %v, %v; ожидали %v", pattern, name, got, err, want)
	}
}

func TestGlobBasic(t *testing.T) {
	chkGlob(t, "*.log", "app.log", true)
	chkGlob(t, "*.log", "app.log.gz", false)
	chkGlob(t, "app-??.txt", "app-01.txt", true)
	chkGlob(t, "app-??.txt", "app-1.txt", false)
	chkGlob(t, "*", "", true)
	chkGlob(t, "", "", true)
	chkGlob(t, "", "a", false)
	chkGlob(t, "?", "", false)
	chkGlob(t, "a*b*c", "a-x-b-y-c", true)
	chkGlob(t, "a*b*c", "a-x-c-y-b", false)
	chkGlob(t, "*ab", "aab", true)
	chkGlob(t, "a*ab", "aXaXab", true)
	chkGlob(t, "logs/*/err", "logs/2024/06/err", true)
	chkGlob(t, "Report*", "report.pdf", false)
}

func TestGlobRunes(t *testing.T) {
	chkGlob(t, "отчёт-?.pdf", "отчёт-я.pdf", true)
	chkGlob(t, "??", "ёж", true)
	chkGlob(t, "???", "ёж", false)
	chkGlob(t, "*ж", "ёж", true)
}

func TestGlobEscape(t *testing.T) {
	chkGlob(t, `\*.txt`, "*.txt", true)
	chkGlob(t, `\*.txt`, "a.txt", false)
	chkGlob(t, `what\?`, "what?", true)
	chkGlob(t, `what\?`, "whatx", false)
	chkGlob(t, `C:\\*`, `C:\dir`, true)
	chkGlob(t, `\a*`, "abc", true)
}

func TestGlobBadPattern(t *testing.T) {
	for _, p := range []string{`abc\`, `x*\`, `\`} {
		ok, err := Match(p, "совсем другое")
		if !errors.Is(err, ErrBadPattern) || ok {
			t.Fatalf("Match(%q, …) = %v, %v; ожидали false и ErrBadPattern", p, ok, err)
		}
	}
}

func TestGlobNoExponent(t *testing.T) {
	pattern := strings.Repeat("a*", 40) + "b"
	name := strings.Repeat("a", 200)
	done := make(chan bool, 1)
	go func() {
		ok, _ := Match(pattern, name)
		done <- ok
	}()
	select {
	case ok := <-done:
		if ok {
			t.Fatalf("Match(%q, 200×\"a\") = true, ожидали false", pattern)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Match на шаблоне a*a*…*b думает дольше 3 секунд — экспоненциальный перебор")
	}
	if ok, _ := Match(strings.Repeat("*a", 50), strings.Repeat("a", 5000)); !ok {
		t.Fatalf("Match(\"*a\"×50, 5000×\"a\") = false, ожидали true")
	}
}
