// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/matching-brackets, MIT, © Exercism). Руками не править.
package main

func TestIsPaired(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"paired square brackets", "[]", true},
		{"empty string", "", true},
		{"unpaired brackets", "[[", false},
		{"wrong ordered brackets", "}{", false},
		{"wrong closing bracket", "{]", false},
		{"paired with whitespace", "{ }", true},
		{"partially paired brackets", "{[])", false},
		{"simple nested brackets", "{[]}", true},
		{"several paired brackets", "{}[]", true},
		{"paired and nested brackets", "([{}({}[])])", true},
		{"unopened closing brackets", "{[)][]}", false},
		{"unpaired and nested brackets", "([{])", false},
		{"paired and wrong nested brackets", "[({]})", false},
		{"paired and wrong nested brackets but innermost are correct", "[({}])", false},
		{"paired and incomplete brackets", "{}[", false},
		{"too many closing brackets", "[]]", false},
		{"early unexpected brackets", ")()", false},
		{"early mismatched brackets", "{)()", false},
		{"math expression", "(((185 + 223.85) * 15) - 543)/2", true},
		{"complex latex expression", "\\left(\\begin{array}{cc} \\frac{1}{3} & x\\\\ \\mathrm{e}^{x} &... x^2 \\end{array}\\right)", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := IsPaired(c.value)
			if got != c.want {
				t.Fatalf("IsPaired(%q) = %v, ожидали %v", c.value, got, c.want)
			}
		})
	}
}
