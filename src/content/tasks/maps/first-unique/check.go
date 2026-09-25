package main

func TestFirstUniqueASCII(t *testing.T) {
	cases := map[string]rune{"leetcode": 'l', "loveleetcode": 'v', "aabbc": 'c', "z": 'z'}
	for s, want := range cases {
		if got, ok := FirstUnique(s); !ok || got != want {
			t.Fatalf("FirstUnique(%q) = %q, %v; ожидали %q", s, got, ok, want)
		}
	}
}

func TestFirstUniqueNone(t *testing.T) {
	for _, s := range []string{"", "aabb", "абаб"} {
		if got, ok := FirstUnique(s); ok {
			t.Fatalf("FirstUnique(%q) = %q, true; уникальных символов нет", s, got)
		}
	}
}

func TestFirstUniqueRunes(t *testing.T) {
	if got, ok := FirstUnique("ааббвг"); !ok || got != 'в' {
		t.Fatalf("FirstUnique(\"ааббвг\") = %q, %v; ожидали 'в'", got, ok)
	}
	// «ё» и «с» делят первый байт в UTF-8: побайтовый подсчёт их спутает.
	if got, ok := FirstUnique("ёсё"); !ok || got != 'с' {
		t.Fatalf("FirstUnique(\"ёсё\") = %q, %v; ожидали 'с'", got, ok)
	}
}

func TestFirstUniqueOrder(t *testing.T) {
	for range 20 {
		if got, _ := FirstUnique("xyzxq"); got != 'y' {
			t.Fatalf("FirstUnique(\"xyzxq\") = %q; ожидали 'y' — первый по порядку в строке", got)
		}
	}
}
