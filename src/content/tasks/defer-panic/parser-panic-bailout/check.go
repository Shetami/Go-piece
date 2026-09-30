package main

func chkVars(name string) int {
	switch name {
	case "x":
		return 5
	case "limit":
		return 100
	}
	panic("неизвестная переменная " + name)
}

func TestEvalValues(t *testing.T) {
	cases := []struct {
		expr string
		want int
	}{
		{"2 + 3 * 4", 14},
		{"10 - 4 - 3", 3},
		{"100 / 7 / 2", 7},
		{"-(2+3) * -2", 10},
		{"x * (x - 1)", 20},
		{"limit/x/(1+1)", 10},
		{"42", 42},
	}
	for _, c := range cases {
		got, err := Eval(c.expr, chkVars)
		if err != nil || got != c.want {
			t.Fatalf("Eval(%q) = %d, %v; ожидали %d, nil", c.expr, got, err, c.want)
		}
	}
}

func TestEvalSyntaxErrors(t *testing.T) {
	cases := []struct {
		expr string
		pos  int
	}{
		{"1 + * 2", 4},
		{"(1 + 2", 6},
		{"2 * (3 + x) )", 12},
		{"", 0},
		{"3 +", 3},
		{"4 $ 4", 2},
	}
	for _, c := range cases {
		_, err := Eval(c.expr, chkVars)
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Fatalf("Eval(%q): ошибка %v, ожидали *ParseError", c.expr, err)
		}
		if pe.Pos != c.pos {
			t.Fatalf("Eval(%q): Pos = %d, ожидали %d", c.expr, pe.Pos, c.pos)
		}
	}
}

func TestEvalDivByZero(t *testing.T) {
	_, err := Eval("7 / (x - 5)", chkVars)
	var pe *ParseError
	if !errors.Is(err, ErrDivByZero) || !errors.As(err, &pe) || pe.Pos != 2 {
		t.Fatalf("Eval(\"7 / (x - 5)\") = %v, ожидали *ParseError на позиции 2 с ErrDivByZero", err)
	}
}

func TestEvalDeepNesting(t *testing.T) {
	const depth = 5000
	ok := strings.Repeat("(", depth) + "1" + strings.Repeat(")", depth)
	if v, err := Eval(ok, chkVars); err != nil || v != 1 {
		t.Fatalf("глубокая вложенность: %d, %v; ожидали 1, nil", v, err)
	}
	bad := strings.Repeat("(", depth) + "1" + strings.Repeat(")", depth-1)
	_, err := Eval(bad, chkVars)
	var pe *ParseError
	if !errors.As(err, &pe) || pe.Pos != len(bad) {
		t.Fatalf("незакрытая скобка на глубине %d: %v, ожидали *ParseError с Pos=%d", depth, err, len(bad))
	}
}

func TestEvalForeignPanic(t *testing.T) {
	for _, val := range []any{"lookup сломался", &ParseError{Pos: 99, Msg: "чужая"}} {
		func() {
			defer func() {
				if r := recover(); r != val {
					t.Fatalf("паника из lookup должна пролететь как есть: ожидали %v, recover() = %v", val, r)
				}
			}()
			v, err := Eval("1 + (2 * y)", func(string) int { panic(val) })
			t.Fatalf("паника из lookup проглочена: Eval вернул %d, %v", v, err)
		}()
	}
}
