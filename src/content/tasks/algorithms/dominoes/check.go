// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/dominoes, MIT, © Exercism). Руками не править.
package main

func TestCanChain(t *testing.T) {
	cases := []struct {
		name     string
		dominoes [][2]int
		want     bool
	}{
		{"empty input = empty output", [][2]int{}, true},
		{"singleton input = singleton output", [][2]int{{1, 1}}, true},
		{"singleton that can't be chained", [][2]int{{1, 2}}, false},
		{"three elements", [][2]int{{1, 2}, {3, 1}, {2, 3}}, true},
		{"can reverse dominoes", [][2]int{{1, 2}, {1, 3}, {2, 3}}, true},
		{"can't be chained", [][2]int{{1, 2}, {4, 1}, {2, 3}}, false},
		{"disconnected - simple", [][2]int{{1, 1}, {2, 2}}, false},
		{"disconnected - double loop", [][2]int{{1, 2}, {2, 1}, {3, 4}, {4, 3}}, false},
		{"disconnected - single isolated", [][2]int{{1, 2}, {2, 3}, {3, 1}, {4, 4}}, false},
		{"need backtrack", [][2]int{{1, 2}, {2, 3}, {3, 1}, {2, 4}, {2, 4}}, true},
		{"separate loops", [][2]int{{1, 2}, {2, 3}, {3, 1}, {1, 1}, {2, 2}, {3, 3}}, true},
		{"nine elements", [][2]int{{1, 2}, {5, 3}, {3, 1}, {1, 2}, {2, 4}, {1, 6}, {2, 3}, {3, 4}, {5, 6}}, true},
		{"separate three-domino loops", [][2]int{{1, 2}, {2, 3}, {3, 1}, {4, 5}, {5, 6}, {6, 4}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CanChain(c.dominoes)
			if got != c.want {
				t.Fatalf("CanChain(%v) = %v, ожидали %v", c.dominoes, got, c.want)
			}
		})
	}
}
