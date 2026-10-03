// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/flower-field, MIT, © Exercism). Руками не править.
package main

func TestAnnotate(t *testing.T) {
	cases := []struct {
		name   string
		garden []string
		want   []string
	}{
		{"no rows", []string{}, []string{}},
		{"no columns", []string{""}, []string{""}},
		{"no flowers", []string{"   ", "   ", "   "}, []string{"   ", "   ", "   "}},
		{"garden full of flowers", []string{"***", "***", "***"}, []string{"***", "***", "***"}},
		{"flower surrounded by spaces", []string{"   ", " * ", "   "}, []string{"111", "1*1", "111"}},
		{"space surrounded by flowers", []string{"***", "* *", "***"}, []string{"***", "*8*", "***"}},
		{"horizontal line", []string{" * * "}, []string{"1*2*1"}},
		{"horizontal line, flowers at edges", []string{"*   *"}, []string{"*1 1*"}},
		{"vertical line", []string{" ", "*", " ", "*", " "}, []string{"1", "*", "2", "*", "1"}},
		{"vertical line, flowers at edges", []string{"*", " ", " ", " ", "*"}, []string{"*", "1", " ", "1", "*"}},
		{"cross", []string{"  *  ", "  *  ", "*****", "  *  ", "  *  "}, []string{" 2*2 ", "25*52", "*****", "25*52", " 2*2 "}},
		{"large garden", []string{" *  * ", "  *   ", "    * ", "   * *", " *  * ", "      "}, []string{"1*22*1", "12*322", " 123*2", "112*4*", "1*22*2", "111111"}},
		{"multiple adjacent flowers", []string{" ** "}, []string{"1**1"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Annotate(c.garden)
			if !sameResult(got, c.want) {
				t.Fatalf("Annotate(%q) = %q, ожидали %q", c.garden, got, c.want)
			}
		})
	}
}

// sameResult — reflect.DeepEqual, который не отличает nil от пустого слайса
// или мапы: для ответа это одно и то же.
func sameResult(got, want any) bool {
	g, w := reflect.ValueOf(got), reflect.ValueOf(want)
	if g.Kind() == w.Kind() && (g.Kind() == reflect.Slice || g.Kind() == reflect.Map) && g.Len() == 0 && w.Len() == 0 {
		return true
	}
	return reflect.DeepEqual(got, want)
}
