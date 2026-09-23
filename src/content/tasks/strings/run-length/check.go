package main

func TestCompressBasic(t *testing.T) {
	cases := map[string]string{
		"aaabcc":   "a3bc2",
		"abc":      "abc",
		"zzzzzzzz": "z8",
		"aabbaa":   "a2b2a2",
	}
	for in, want := range cases {
		if got := Compress(in); got != want {
			t.Fatalf("Compress(%q) = %q, ожидали %q", in, got, want)
		}
	}
}

func TestCompressRunes(t *testing.T) {
	if got := Compress("ёёёжж"); got != "ё3ж2" {
		t.Fatalf(`Compress("ёёёжж") = %q, ожидали "ё3ж2"`, got)
	}
}

func TestCompressLongRun(t *testing.T) {
	in := strings.Repeat("x", 12) + "y"
	if got := Compress(in); got != "x12y" {
		t.Fatalf("12 одинаковых подряд: %q, ожидали \"x12y\"", got)
	}
}

func TestCompressEmpty(t *testing.T) {
	if got := Compress("a"); got != "a" {
		t.Fatalf(`Compress("a") = %q`, got)
	}
	if got := Compress(""); got != "" {
		t.Fatalf(`Compress("") = %q, ожидали пустую строку`, got)
	}
}
