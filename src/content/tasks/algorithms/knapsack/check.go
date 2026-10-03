// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/knapsack, MIT, © Exercism). Руками не править.
package main

func TestMaxValue(t *testing.T) {
	cases := []struct {
		name          string
		maximumWeight int
		items         []Item
		want          int
	}{
		{"no items", 100, []Item{}, 0},
		{"one item, too heavy", 10, []Item{{Weight: 100, Value: 1}}, 0},
		{"five items (cannot be greedy by weight)", 10, []Item{{Weight: 2, Value: 5}, {Weight: 2, Value: 5}, {Weight: 2, Value: 5}, {Weight: 2, Value: 5}, {Weight: 10, Value: 21}}, 21},
		{"five items (cannot be greedy by value)", 10, []Item{{Weight: 2, Value: 20}, {Weight: 2, Value: 20}, {Weight: 2, Value: 20}, {Weight: 2, Value: 20}, {Weight: 10, Value: 50}}, 80},
		{"example knapsack", 10, []Item{{Weight: 5, Value: 10}, {Weight: 4, Value: 40}, {Weight: 6, Value: 30}, {Weight: 4, Value: 50}}, 90},
		{"8 items", 104, []Item{{Weight: 25, Value: 350}, {Weight: 35, Value: 400}, {Weight: 45, Value: 450}, {Weight: 5, Value: 20}, {Weight: 25, Value: 70}, {Weight: 3, Value: 8}, {Weight: 2, Value: 5}, {Weight: 2, Value: 5}}, 900},
		{"15 items", 750, []Item{{Weight: 70, Value: 135}, {Weight: 73, Value: 139}, {Weight: 77, Value: 149}, {Weight: 80, Value: 150}, {Weight: 82, Value: 156}, {Weight: 87, Value: 163}, {Weight: 90, Value: 173}, {Weight: 94, Value: 184}, {Weight: 98, Value: 192}, {Weight: 106, Value: 201}, {Weight: 110, Value: 210}, {Weight: 113, Value: 214}, {Weight: 115, Value: 221}, {Weight: 118, Value: 229}, {Weight: 120, Value: 240}}, 1458},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := MaxValue(c.maximumWeight, c.items)
			if got != c.want {
				t.Fatalf("MaxValue(%v, %v) = %v, ожидали %v", c.maximumWeight, c.items, got, c.want)
			}
		})
	}
}
