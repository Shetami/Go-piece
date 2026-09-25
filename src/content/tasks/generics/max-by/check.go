package main

type employee struct {
	name   string
	salary int
}

func TestMaxByStruct(t *testing.T) {
	staff := []employee{{"аня", 100}, {"боря", 300}, {"вика", 200}}
	got, ok := MaxBy(staff, func(e employee) int { return e.salary })
	if !ok || got.name != "боря" {
		t.Fatalf("получили %v, %v; ожидали боря", got, ok)
	}
}

func TestMaxByStringKey(t *testing.T) {
	got, _ := MaxBy([]string{"go", "rust", "c"}, func(s string) string { return s })
	if got != "rust" {
		t.Fatalf("получили %q, ожидали rust", got)
	}
	longest, _ := MaxBy([]string{"go", "rust", "zig", "odin"}, func(s string) int { return len(s) })
	if longest != "rust" {
		t.Fatalf("самая длинная — первая из равных, rust; получили %q", longest)
	}
}

func TestMaxByEmpty(t *testing.T) {
	got, ok := MaxBy([]employee(nil), func(e employee) int { return e.salary })
	if ok || got != (employee{}) {
		t.Fatalf("пустой вход: %v, %v", got, ok)
	}
}

func TestMaxByNegative(t *testing.T) {
	got, _ := MaxBy([]int{-5, -2, -9}, func(x int) int { return x })
	if got != -2 {
		t.Fatalf("отрицательные ключи: получили %d, ожидали -2", got)
	}
}

func TestMaxByKeyCalls(t *testing.T) {
	calls := 0
	MaxBy([]int{1, 2, 3, 4, 5}, func(x int) float64 { calls++; return float64(x) })
	if calls > 5 {
		t.Fatalf("key вызвана %d раз на 5 элементов", calls)
	}
}
