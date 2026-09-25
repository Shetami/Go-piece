package main

func TestExprEval(t *testing.T) {
	// (2 + 3) * -4
	var e Expr = Mul{Add{Num{2}, Num{3}}, Neg{Num{4}}}
	if got := e.Eval(); got != -20 {
		t.Fatalf("Eval = %v, ожидали -20", got)
	}
	if got := e.String(); got != "((2 + 3) * -4)" {
		t.Fatalf("String = %q, ожидали %q", got, "((2 + 3) * -4)")
	}
}

func TestExprNumbers(t *testing.T) {
	if s := (Num{0.5}).String(); s != "0.5" {
		t.Fatalf("Num{0.5}.String() = %q", s)
	}
	if s := (Num{100}).String(); s != "100" {
		t.Fatalf("Num{100}.String() = %q", s)
	}
}

func TestExprPrintf(t *testing.T) {
	var e Expr = Add{Num{1}, Mul{Num{2}, Num{3}}}
	if got := fmt.Sprint(e); got != "(1 + (2 * 3))" {
		t.Fatalf("fmt.Sprint(e) = %q — fmt должен вызвать String", got)
	}
	if e.Eval() != 7 {
		t.Fatalf("Eval = %v, ожидали 7", e.Eval())
	}
}

func TestExprPointers(t *testing.T) {
	var e Expr = &Neg{&Num{3}}
	if e.Eval() != -3 || e.String() != "-3" {
		t.Fatalf("узлы по указателю тоже должны работать: %v %q", e.Eval(), e.String())
	}
}
