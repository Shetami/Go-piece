package main

func TestReduceSum(t *testing.T) {
	if got := Reduce([]int{1, 2, 3, 4}, 0, func(a, x int) int { return a + x }); got != 10 {
		t.Fatalf("сумма = %d, ожидали 10", got)
	}
}

func TestReduceDifferentType(t *testing.T) {
	got := Reduce([]string{"go", "rust", "c"}, 0, func(a int, s string) int { return a + len(s) })
	if got != 7 {
		t.Fatalf("сумма длин = %d, ожидали 7", got)
	}
}

func TestReduceOrder(t *testing.T) {
	got := Reduce([]string{"a", "b", "c"}, ">", func(a string, s string) string { return a + s })
	if got != ">abc" {
		t.Fatalf("свёртка слева направо: %q, ожидали \">abc\"", got)
	}
}

func TestReduceIntoMap(t *testing.T) {
	counts := Reduce([]string{"x", "y", "x"}, map[string]int{}, func(m map[string]int, s string) map[string]int {
		m[s]++
		return m
	})
	if !reflect.DeepEqual(counts, map[string]int{"x": 2, "y": 1}) {
		t.Fatalf("подсчёт через Reduce: %v", counts)
	}
}
