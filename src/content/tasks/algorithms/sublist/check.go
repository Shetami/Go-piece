// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/sublist, MIT, © Exercism). Руками не править.
package main

func TestCompare(t *testing.T) {
	cases := []struct {
		name    string
		listOne []int
		listTwo []int
		want    string
	}{
		{"empty lists", []int{}, []int{}, "equal"},
		{"empty list within non empty list", []int{}, []int{1, 2, 3}, "sublist"},
		{"non empty list contains empty list", []int{1, 2, 3}, []int{}, "superlist"},
		{"list equals itself", []int{1, 2, 3}, []int{1, 2, 3}, "equal"},
		{"different lists", []int{1, 2, 3}, []int{2, 3, 4}, "unequal"},
		{"false start", []int{1, 2, 5}, []int{0, 1, 2, 3, 1, 2, 5, 6}, "sublist"},
		{"consecutive", []int{1, 1, 2}, []int{0, 1, 1, 1, 2, 1, 2}, "sublist"},
		{"sublist at start", []int{0, 1, 2}, []int{0, 1, 2, 3, 4, 5}, "sublist"},
		{"sublist in middle", []int{2, 3, 4}, []int{0, 1, 2, 3, 4, 5}, "sublist"},
		{"sublist at end", []int{3, 4, 5}, []int{0, 1, 2, 3, 4, 5}, "sublist"},
		{"at start of superlist", []int{0, 1, 2, 3, 4, 5}, []int{0, 1, 2}, "superlist"},
		{"in middle of superlist", []int{0, 1, 2, 3, 4, 5}, []int{2, 3}, "superlist"},
		{"at end of superlist", []int{0, 1, 2, 3, 4, 5}, []int{3, 4, 5}, "superlist"},
		{"first list missing element from second list", []int{1, 3}, []int{1, 2, 3}, "unequal"},
		{"second list missing element from first list", []int{1, 2, 3}, []int{1, 3}, "unequal"},
		{"first list missing additional digits from second list", []int{1, 2}, []int{1, 22}, "unequal"},
		{"order matters to a list", []int{1, 2, 3}, []int{3, 2, 1}, "unequal"},
		{"same digits but different numbers", []int{1, 0, 1}, []int{10, 1}, "unequal"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Compare(c.listOne, c.listTwo)
			if got != c.want {
				t.Fatalf("Compare(%v, %v) = %q, ожидали %q", c.listOne, c.listTwo, got, c.want)
			}
		})
	}
}
