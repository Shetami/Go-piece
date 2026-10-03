// Сгенерировано scripts/exercism/generate.ts из exercism/problem-specifications
// (exercises/all-your-base, MIT, © Exercism). Руками не править.
package main

func TestRebase(t *testing.T) {
	cases := []struct {
		name       string
		inputBase  int
		digits     []int
		outputBase int
		want       []int
		wantErr    bool
	}{
		{"single bit one to decimal", 2, []int{1}, 10, []int{1}, false},
		{"binary to single decimal", 2, []int{1, 0, 1}, 10, []int{5}, false},
		{"single decimal to binary", 10, []int{5}, 2, []int{1, 0, 1}, false},
		{"binary to multiple decimal", 2, []int{1, 0, 1, 0, 1, 0}, 10, []int{4, 2}, false},
		{"decimal to binary", 10, []int{4, 2}, 2, []int{1, 0, 1, 0, 1, 0}, false},
		{"trinary to hexadecimal", 3, []int{1, 1, 2, 0}, 16, []int{2, 10}, false},
		{"hexadecimal to trinary", 16, []int{2, 10}, 3, []int{1, 1, 2, 0}, false},
		{"15-bit integer", 97, []int{3, 46, 60}, 73, []int{6, 10, 45}, false},
		{"empty list", 2, []int{}, 10, []int{0}, false},
		{"single zero", 10, []int{0}, 2, []int{0}, false},
		{"multiple zeros", 10, []int{0, 0, 0}, 2, []int{0}, false},
		{"leading zeros", 7, []int{0, 6, 0}, 10, []int{4, 2}, false},
		{"input base is one", 1, []int{0}, 10, nil, true},
		{"input base is zero", 0, []int{}, 10, nil, true},
		{"input base is negative", -2, []int{1}, 10, nil, true},
		{"negative digit", 2, []int{1, -1, 1, 0, 1, 0}, 10, nil, true},
		{"invalid positive digit", 2, []int{1, 2, 1, 0, 1, 0}, 10, nil, true},
		{"output base is one", 2, []int{1, 0, 1, 0, 1, 0}, 1, nil, true},
		{"output base is zero", 10, []int{7}, 0, nil, true},
		{"output base is negative", 2, []int{1}, -7, nil, true},
		{"both bases are negative", -2, []int{1}, -7, nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Rebase(c.inputBase, c.digits, c.outputBase)
			if c.wantErr {
				if err == nil {
					t.Fatalf("Rebase(%v, %v, %v) = %v, ожидали ошибку", c.inputBase, c.digits, c.outputBase, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Rebase(%v, %v, %v): неожиданная ошибка %v", c.inputBase, c.digits, c.outputBase, err)
			}
			if !sameResult(got, c.want) {
				t.Fatalf("Rebase(%v, %v, %v) = %v, ожидали %v", c.inputBase, c.digits, c.outputBase, got, c.want)
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
