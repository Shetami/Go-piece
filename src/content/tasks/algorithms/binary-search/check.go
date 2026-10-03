// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/binary-search, MIT, © Exercism). Руками не править.
package main

func TestSearch(t *testing.T) {
	cases := []struct {
		name  string
		array []int
		value int
		want  int
	}{
		{"finds a value in an array with one element", []int{6}, 6, 0},
		{"finds a value in the middle of an array", []int{1, 3, 4, 6, 8, 9, 11}, 6, 3},
		{"finds a value at the beginning of an array", []int{1, 3, 4, 6, 8, 9, 11}, 1, 0},
		{"finds a value at the end of an array", []int{1, 3, 4, 6, 8, 9, 11}, 11, 6},
		{"finds a value in an array of odd length", []int{1, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 634}, 144, 9},
		{"finds a value in an array of even length", []int{1, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377}, 21, 5},
		{"identifies that a value is not included in the array", []int{1, 3, 4, 6, 8, 9, 11}, 7, -1},
		{"a value smaller than the array's smallest value is not found", []int{1, 3, 4, 6, 8, 9, 11}, 0, -1},
		{"a value larger than the array's largest value is not found", []int{1, 3, 4, 6, 8, 9, 11}, 13, -1},
		{"nothing is found in an empty array", []int{}, 1, -1},
		{"nothing is found when the left and right bounds cross", []int{1, 2}, 0, -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Search(c.array, c.value)
			if got != c.want {
				t.Fatalf("Search(%v, %v) = %v, ожидали %v", c.array, c.value, got, c.want)
			}
		})
	}
}
