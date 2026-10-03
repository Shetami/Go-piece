// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/run-length-encoding, MIT, © Exercism). Руками не править.
package main

func TestEncode(t *testing.T) {
	cases := []struct {
		name   string
		string string
		want   string
	}{
		{"empty string", "", ""},
		{"single characters only are encoded without count", "XYZ", "XYZ"},
		{"string with no single characters", "AABBBCCCC", "2A3B4C"},
		{"single characters mixed with repeated characters", "WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWWWWWWWWWWWWB", "12WB12W3B24WB"},
		{"multiple whitespace mixed in string", "  hsqq qww  ", "2 hs2q q2w2 "},
		{"lowercase characters", "aabbbcccc", "2a3b4c"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Encode(c.string)
			if got != c.want {
				t.Fatalf("Encode(%q) = %q, ожидали %q", c.string, got, c.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	cases := []struct {
		name   string
		string string
		want   string
	}{
		{"empty string", "", ""},
		{"single characters only", "XYZ", "XYZ"},
		{"string with no single characters", "2A3B4C", "AABBBCCCC"},
		{"single characters with repeated characters", "12WB12W3B24WB", "WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWWWWWWWWWWWWB"},
		{"multiple whitespace mixed in string", "2 hs2q q2w2 ", "  hsqq qww  "},
		{"lowercase string", "2a3b4c", "aabbbcccc"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Decode(c.string)
			if got != c.want {
				t.Fatalf("Decode(%q) = %q, ожидали %q", c.string, got, c.want)
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	cases := []struct {
		name   string
		string string
		want   string
	}{
		{"encode followed by decode gives original string", "zzz ZZ  zZ", "zzz ZZ  zZ"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Decode(Encode(c.string))
			if got != c.want {
				t.Fatalf("Decode(Encode(%q)) = %q, ожидали %q", c.string, got, c.want)
			}
		})
	}
}
