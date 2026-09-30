package main

func TestCompileEval(t *testing.T) {
	cases := []struct {
		src  string
		vars map[string]float64
		want float64
	}{
		{"2 3 +", nil, 5},
		{"10 4 -", nil, 6},
		{"1 8 /", nil, 0.125},
		{"price qty * discount -", map[string]float64{"price": 2.5, "qty": 4, "discount": 3}, 7},
		{"-3 x_1 *", map[string]float64{"x_1": 2}, -6},
		{"a b c + * 2 /", map[string]float64{"a": 3, "b": 1, "c": 5}, 9},
		{"42", nil, 42},
	}
	for _, c := range cases {
		e, err := Compile(c.src)
		if err != nil {
			t.Fatalf("Compile(%q): %v", c.src, err)
		}
		if got, err := e(c.vars); err != nil || got != c.want {
			t.Fatalf("%q = %v, %v; ожидали %v", c.src, got, err, c.want)
		}
	}
}

func TestCompileSyntaxErrors(t *testing.T) {
	for _, src := range []string{"", "   ", "1 +", "+", "1 2", "a b c +", "2 3 %", "1x 2 +", "2 3 + *"} {
		e, err := Compile(src)
		if !errors.Is(err, ErrSyntax) {
			t.Fatalf("Compile(%q): ожидали ErrSyntax ещё при компиляции, получили %v", src, err)
		}
		if e != nil {
			t.Fatalf("Compile(%q) с ошибкой вернул не nil Expr", src)
		}
	}
}

func TestCompileRuntimeErrors(t *testing.T) {
	e, err := Compile("total count /")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if _, err := e(map[string]float64{"total": 1}); !errors.Is(err, ErrUnknown) || !strings.Contains(err.Error(), "count") {
		t.Fatalf("нет переменной count: ожидали ErrUnknown с именем, получили %v", err)
	}
	if _, err := e(map[string]float64{"total": 1, "count": 0}); !errors.Is(err, ErrDivZero) {
		t.Fatalf("деление на ноль: ожидали ErrDivZero, получили %v", err)
	}
	if v, err := e(map[string]float64{"total": 9, "count": 3}); err != nil || v != 3 {
		t.Fatalf("то же выражение на других переменных: %v, %v; ожидали 3", v, err)
	}
}

func TestCompileOnce(t *testing.T) {
	e, err := Compile("a 2 * b + c d - /")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	vars := map[string]float64{"a": 1, "b": 2, "c": 5, "d": 3}
	allocs := testing.AllocsPerRun(100, func() {
		if v, _ := e(vars); v != 2 {
			panic("неверный результат")
		}
	})
	if allocs > 0 {
		t.Fatalf("вычисление выделяет память (%.0f раз) — похоже, строка разбирается при каждом вызове, а не в Compile", allocs)
	}
}
