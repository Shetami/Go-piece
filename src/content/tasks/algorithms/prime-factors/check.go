// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/prime-factors, MIT, © Exercism). Руками не править.
package main

func TestFactors(t *testing.T) {
	cases := []struct {
		name  string
		value int
		want  []int
	}{
		{"no factors", 1, []int{}},
		{"prime number", 2, []int{2}},
		{"another prime number", 3, []int{3}},
		{"square of a prime", 9, []int{3, 3}},
		{"product of first prime", 4, []int{2, 2}},
		{"cube of a prime", 8, []int{2, 2, 2}},
		{"product of second prime", 27, []int{3, 3, 3}},
		{"product of third prime", 625, []int{5, 5, 5, 5}},
		{"product of first and second prime", 6, []int{2, 3}},
		{"product of primes and non-primes", 12, []int{2, 2, 3}},
		{"product of primes", 901255, []int{5, 17, 23, 461}},
		{"factors include a large prime", 93819012551, []int{11, 9539, 894119}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Factors(c.value)
			if !sameResult(got, c.want) {
				t.Fatalf("Factors(%v) = %v, ожидали %v", c.value, got, c.want)
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
