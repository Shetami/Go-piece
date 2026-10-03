// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/pascals-triangle, MIT, © Exercism). Руками не править.
package main

func TestPascal(t *testing.T) {
	cases := []struct {
		name  string
		count int
		want  [][]int
	}{
		{"zero rows", 0, [][]int{}},
		{"single row", 1, [][]int{{1}}},
		{"two rows", 2, [][]int{{1}, {1, 1}}},
		{"three rows", 3, [][]int{{1}, {1, 1}, {1, 2, 1}}},
		{"four rows", 4, [][]int{{1}, {1, 1}, {1, 2, 1}, {1, 3, 3, 1}}},
		{"five rows", 5, [][]int{{1}, {1, 1}, {1, 2, 1}, {1, 3, 3, 1}, {1, 4, 6, 4, 1}}},
		{"six rows", 6, [][]int{{1}, {1, 1}, {1, 2, 1}, {1, 3, 3, 1}, {1, 4, 6, 4, 1}, {1, 5, 10, 10, 5, 1}}},
		{"ten rows", 10, [][]int{{1}, {1, 1}, {1, 2, 1}, {1, 3, 3, 1}, {1, 4, 6, 4, 1}, {1, 5, 10, 10, 5, 1}, {1, 6, 15, 20, 15, 6, 1}, {1, 7, 21, 35, 35, 21, 7, 1}, {1, 8, 28, 56, 70, 56, 28, 8, 1}, {1, 9, 36, 84, 126, 126, 84, 36, 9, 1}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Pascal(c.count)
			if !sameResult(got, c.want) {
				t.Fatalf("Pascal(%v) = %v, ожидали %v", c.count, got, c.want)
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
