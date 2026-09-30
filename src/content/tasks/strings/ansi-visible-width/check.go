package main

const (
	chkRed   = "\x1b[31m"
	chkBold  = "\x1b[1;4m"
	chkReset = "\x1b[0m"
	chkLink  = "\x1b]8;;https://go.dev/doc\x1b\\"
	chkEnd   = "\x1b]8;;\x1b\\"
)

func TestAnsiWidth(t *testing.T) {
	cases := map[string]int{
		"":                           0,
		"plain":                      5,
		chkRed + "ошибка" + chkReset: 6,
		chkBold + "OK" + chkReset + " готово": 9,
		"\x1b[2K\x1b[1Gпрогресс 50%":          12,
		chkLink + "docs" + chkEnd:             4,
		"\x1b]0;заголовок окна\x07текст":      5,
		"a\x1b7b\x1b8c":                       3, // ESC 7 и ESC 8 — двухсимвольные последовательности
		"хвост\x1b[31":                        5, // оборванная последовательность
		"\x1b]8;;http://x":                    0,
	}
	for s, want := range cases {
		if got := VisibleWidth(s); got != want {
			t.Fatalf("VisibleWidth(%q) = %d, ожидали %d", s, got, want)
		}
	}
}

func TestAnsiPad(t *testing.T) {
	s := chkRed + "ок" + chkReset
	if got := PadRight(s, 5); got != s+"   " {
		t.Fatalf("PadRight(%q, 5) = %q, ожидали %q", s, got, s+"   ")
	}
	if got := PadRight("длинная", 3); got != "длинная" {
		t.Fatalf("PadRight(\"длинная\", 3) = %q", got)
	}
}

func chkTV(t *testing.T, s string, w int, want string) {
	t.Helper()
	if got := TruncateVisible(s, w); got != want {
		t.Fatalf("TruncateVisible(%q, %d) = %q, ожидали %q", s, w, got, want)
	}
}

func TestAnsiTruncate(t *testing.T) {
	chkTV(t, "привет", 3, "при")
	chkTV(t, "привет", 10, "привет")
	s := chkRed + "ошибка" + chkReset
	chkTV(t, s, 6, s)
	chkTV(t, s, 3, chkRed+"оши"+chkReset)
	chkTV(t, s, 0, chkRed+chkReset)
	chkTV(t, chkRed+"ab"+chkReset+"cd", 2, chkRed+"ab"+chkReset+chkReset)
	chkTV(t, "ab"+chkRed+"cd"+chkReset, 2, "ab"+chkRed+chkReset)
	chkTV(t, "abc"+chkRed+"d", 1, "a")
	chkTV(t, chkLink+"documentation"+chkEnd, 3, chkLink+"doc"+chkReset)
}
