package main

func chkTrunc(t *testing.T, s string, limit int, want string) {
	t.Helper()
	got := TruncateBytes(s, limit)
	if got != want {
		t.Fatalf("TruncateBytes(%q, %d) = %q, ожидали %q", s, limit, got, want)
	}
	if len(got) > limit && got != s {
		t.Fatalf("TruncateBytes(%q, %d) = %q: %d байт — больше лимита", s, limit, got, len(got))
	}
}

func TestTruncFits(t *testing.T) {
	chkTrunc(t, "Привет", 12, "Привет")
	chkTrunc(t, "", 0, "")
	chkTrunc(t, "hello world", 8, "hello…")
	chkTrunc(t, "hello", 3, "…")
	chkTrunc(t, "hello", 2, "")
}

func TestTruncRunes(t *testing.T) {
	// "Привет" — 12 байт. В лимит 10 помещается 7 байт текста = 3 буквы и полбуквы.
	chkTrunc(t, "Привет", 10, "При…")
	chkTrunc(t, "Привет", 11, "Прив…")
	for limit := 0; limit <= 40; limit++ {
		got := TruncateBytes("Съешь же ещё этих мягких булок", limit)
		if !utf8.ValidString(got) || len(got) > limit {
			t.Fatalf("TruncateBytes(…, %d) = %q: битый UTF-8 или длиннее лимита", limit, got)
		}
	}
}

func TestTruncCombining(t *testing.T) {
	// «й» из двух рун: «и» (2 байта) + U+0306 (2 байта). Отрывать кратку нельзя.
	s := "мой дом"
	chkTrunc(t, s, 9, "мо…") // 6 байт под текст: «мо» + «и» без кратки не годится
	chkTrunc(t, s, 11, "мой…")
	chkTrunc(t, "café au lait", 8, "caf…")
}

func TestTruncEmoji(t *testing.T) {
	family := "👨‍👩‍👧" // 18 байт, один кластер
	chkTrunc(t, "ok "+family+" ok", 20, "ok…")
	chkTrunc(t, "ok "+family+" okay", 24, "ok "+family+"…")
	chkTrunc(t, "hi ❤️!!", 8, "hi…")
	chkTrunc(t, "a👍\U0001F3FDb", 9, "a…")
}

func TestTruncSpacesAndInvalid(t *testing.T) {
	chkTrunc(t, "слово   ещё", 16, "слово…")
	chkTrunc(t, "ab\xffcdef", 6, "ab\xff…")
}
