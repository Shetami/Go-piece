// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/spiral-matrix, MIT, © Exercism). Руками не править.
package main

func TestSpiral(t *testing.T) {
	cases := []struct {
		name string
		size int
		want [][]int
	}{
		{"empty spiral", 0, [][]int{}},
		{"trivial spiral", 1, [][]int{{1}}},
		{"spiral of size 2", 2, [][]int{{1, 2}, {4, 3}}},
		{"spiral of size 3", 3, [][]int{{1, 2, 3}, {8, 9, 4}, {7, 6, 5}}},
		{"spiral of size 4", 4, [][]int{{1, 2, 3, 4}, {12, 13, 14, 5}, {11, 16, 15, 6}, {10, 9, 8, 7}}},
		{"spiral of size 5", 5, [][]int{{1, 2, 3, 4, 5}, {16, 17, 18, 19, 6}, {15, 24, 25, 20, 7}, {14, 23, 22, 21, 8}, {13, 12, 11, 10, 9}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Spiral(c.size)
			if !sameResult(got, c.want) {
				t.Fatalf("Spiral(%v) = %v, ожидали %v", c.size, got, c.want)
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
