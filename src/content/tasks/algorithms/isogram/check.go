// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/isogram, MIT, © Exercism). Руками не править.
package main

func TestIsIsogram(t *testing.T) {
	cases := []struct {
		name   string
		phrase string
		want   bool
	}{
		{"empty string", "", true},
		{"isogram with only lower case characters", "isogram", true},
		{"word with one duplicated character", "eleven", false},
		{"word with one duplicated character from the end of the alphabet", "zzyzx", false},
		{"longest reported english isogram", "subdermatoglyphic", true},
		{"word with duplicated character in mixed case", "Alphabet", false},
		{"word with duplicated character in mixed case, lowercase first", "alphAbet", false},
		{"hypothetical isogrammic word with hyphen", "thumbscrew-japingly", true},
		{"hypothetical word with duplicated character following hyphen", "thumbscrew-jappingly", false},
		{"isogram with duplicated hyphen", "six-year-old", true},
		{"made-up name that is an isogram", "Emily Jung Schwartzkopf", true},
		{"duplicated character in the middle", "accentor", false},
		{"same first and last characters", "angola", false},
		{"word with duplicated character and with two hyphens", "up-to-date", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := IsIsogram(c.phrase)
			if got != c.want {
				t.Fatalf("IsIsogram(%q) = %v, ожидали %v", c.phrase, got, c.want)
			}
		})
	}
}
