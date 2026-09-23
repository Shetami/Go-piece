package main

func TestConcatResult(t *testing.T) {
	if got := Concat([]string{"go", "is", "fun"}, ", "); got != "go, is, fun" {
		t.Fatalf("Concat = %q", got)
	}
	if got := Concat(nil, "-"); got != "" {
		t.Fatalf("Concat(nil) = %q", got)
	}
	if got := Concat([]string{"один"}, "-"); got != "один" {
		t.Fatalf("один элемент: %q", got)
	}
}

func TestConcatOneAllocation(t *testing.T) {
	parts := make([]string, 50)
	for i := range parts {
		parts[i] = strings.Repeat("x", i)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = Concat(parts, "--")
	})
	if allocs > 1 {
		t.Fatalf("Concat выделяет память %.0f раз на вызов, ожидали не больше 1", allocs)
	}
}
