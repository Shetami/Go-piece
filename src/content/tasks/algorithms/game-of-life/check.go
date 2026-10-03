// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/game-of-life, MIT, © Exercism). Руками не править.
package main

func TestTick(t *testing.T) {
	cases := []struct {
		name   string
		matrix [][]int
		want   [][]int
	}{
		{"empty matrix", [][]int{}, [][]int{}},
		{"live cells with zero live neighbors die", [][]int{{0, 0, 0}, {0, 1, 0}, {0, 0, 0}}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}},
		{"live cells with only one live neighbor die", [][]int{{0, 0, 0}, {0, 1, 0}, {0, 1, 0}}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}},
		{"live cells with two live neighbors stay alive", [][]int{{1, 0, 1}, {1, 0, 1}, {1, 0, 1}}, [][]int{{0, 0, 0}, {1, 0, 1}, {0, 0, 0}}},
		{"live cells with three live neighbors stay alive", [][]int{{0, 1, 0}, {1, 0, 0}, {1, 1, 0}}, [][]int{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}}},
		{"dead cells with three live neighbors become alive", [][]int{{1, 1, 0}, {0, 0, 0}, {1, 0, 0}}, [][]int{{0, 0, 0}, {1, 1, 0}, {0, 0, 0}}},
		{"live cells with four or more neighbors die", [][]int{{1, 1, 1}, {1, 1, 1}, {1, 1, 1}}, [][]int{{1, 0, 1}, {0, 0, 0}, {1, 0, 1}}},
		{"bigger matrix", [][]int{{1, 1, 0, 1, 1, 0, 0, 0}, {1, 0, 1, 1, 0, 0, 0, 0}, {1, 1, 1, 0, 0, 1, 1, 1}, {0, 0, 0, 0, 0, 1, 1, 0}, {1, 0, 0, 0, 1, 1, 0, 0}, {1, 1, 0, 0, 0, 1, 1, 1}, {0, 0, 1, 0, 1, 0, 0, 1}, {1, 0, 0, 0, 0, 0, 1, 1}}, [][]int{{1, 1, 0, 1, 1, 0, 0, 0}, {0, 0, 0, 0, 0, 1, 1, 0}, {1, 0, 1, 1, 1, 1, 0, 1}, {1, 0, 0, 0, 0, 0, 0, 1}, {1, 1, 0, 0, 1, 0, 0, 1}, {1, 1, 0, 1, 0, 0, 0, 1}, {1, 0, 0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0, 1, 1}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Tick(c.matrix)
			if !sameResult(got, c.want) {
				t.Fatalf("Tick(%v) = %v, ожидали %v", c.matrix, got, c.want)
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
