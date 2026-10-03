// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/change, MIT, © Exercism). Руками не править.
package main

func TestFewestCoins(t *testing.T) {
	cases := []struct {
		name    string
		coins   []int
		target  int
		want    []int
		wantErr bool
	}{
		{"change for 1 cent", []int{1, 5, 10, 25}, 1, []int{1}, false},
		{"single coin change", []int{1, 5, 10, 25, 100}, 25, []int{25}, false},
		{"multiple coin change", []int{1, 5, 10, 25, 100}, 15, []int{5, 10}, false},
		{"change with Lilliputian Coins", []int{1, 4, 15, 20, 50}, 23, []int{4, 4, 15}, false},
		{"change with Lower Elbonia Coins", []int{1, 5, 10, 21, 25}, 63, []int{21, 21, 21}, false},
		{"large target values", []int{1, 2, 5, 10, 20, 50, 100}, 999, []int{2, 2, 5, 20, 20, 50, 100, 100, 100, 100, 100, 100, 100, 100, 100}, false},
		{"possible change without unit coins available", []int{2, 5, 10, 20, 50}, 21, []int{2, 2, 2, 5, 10}, false},
		{"another possible change without unit coins available", []int{4, 5}, 27, []int{4, 4, 4, 5, 5, 5}, false},
		{"a greedy approach is not optimal", []int{1, 10, 11}, 20, []int{10, 10}, false},
		{"no coins make 0 change", []int{1, 5, 10, 21, 25}, 0, []int{}, false},
		{"error testing for change smaller than the smallest of coins", []int{5, 10}, 3, nil, true},
		{"error if no combination can add up to target", []int{5, 10}, 94, nil, true},
		{"cannot find negative change values", []int{1, 2, 5}, -5, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := FewestCoins(c.coins, c.target)
			if c.wantErr {
				if err == nil {
					t.Fatalf("FewestCoins(%v, %v) = %v, ожидали ошибку", c.coins, c.target, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("FewestCoins(%v, %v): неожиданная ошибка %v", c.coins, c.target, err)
			}
			if !sameResult(got, c.want) {
				t.Fatalf("FewestCoins(%v, %v) = %v, ожидали %v", c.coins, c.target, got, c.want)
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
