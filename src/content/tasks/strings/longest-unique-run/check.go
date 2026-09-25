package main

func TestLongestUniqueClassic(t *testing.T) {
	cases := map[string]int{
		"abcabcbb": 3,
		"bbbbb":    1,
		"pwwkew":   3,
		"":         0,
		"a":        1,
		"abba":     2,
		"dvdf":     3,
		"tmmzuxt":  5,
	}
	for s, want := range cases {
		if got := LongestUnique(s); got != want {
			t.Fatalf("LongestUnique(%q) = %d, ожидали %d", s, got, want)
		}
	}
}

func TestLongestUniqueRunes(t *testing.T) {
	if got := LongestUnique("привет"); got != 6 {
		t.Fatalf("LongestUnique(\"привет\") = %d, ожидали 6 символов, а не байтов", got)
	}
	if got := LongestUnique("ёсё"); got != 2 {
		t.Fatalf("LongestUnique(\"ёсё\") = %d, ожидали 2", got)
	}
}

func TestLongestUniqueLinear(t *testing.T) {
	var b strings.Builder
	for range 20000 {
		b.WriteString("abcdefghijklmnopqrstuvwxyz")
	}
	start := time.Now()
	if got := LongestUnique(b.String()); got != 26 {
		t.Fatalf("получили %d, ожидали 26", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("полмиллиона символов заняли %v — похоже на O(n²)", d)
	}
}
