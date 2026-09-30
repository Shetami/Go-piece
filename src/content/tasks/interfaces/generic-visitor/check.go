package main

// chkCount считает посещённые узлы и падает на переменной "bad".
type chkCount struct{ seen *[]string }

func (c chkCount) Num(n Num) (int, error) { *c.seen = append(*c.seen, "num"); return 1, nil }
func (c chkCount) Var(v Var) (int, error) {
	*c.seen = append(*c.seen, v.Name)
	if v.Name == "bad" {
		return 0, errors.New("stop")
	}
	return 1, nil
}
func (c chkCount) Bin(_ byte, l, r int) (int, error) { return l + r + 1, nil }

type chkAlien struct{}

func (chkAlien) node() {}

func TestVisitorEval(t *testing.T) {
	// (x + 2) * (y - 0.5)
	e := Bin{'*', Bin{'+', Var{"x"}, Num{2}}, &Bin{'-', &Var{"y"}, &Num{0.5}}}
	got, err := Walk[float64](e, Eval{Env: map[string]float64{"x": 1, "y": 3}})
	if err != nil || got != 7.5 {
		t.Fatalf("Eval = %v, %v; ожидали 7.5 (узлы по указателю тоже работают)", got, err)
	}
	s, err := Walk[string](e, Printer{})
	if err != nil || s != "((x + 2) * (y - 0.5))" {
		t.Fatalf("Printer = %q, %v; ожидали \"((x + 2) * (y - 0.5))\"", s, err)
	}
}

func TestVisitorErrors(t *testing.T) {
	_, err := Walk[float64](Bin{'+', Num{1}, Var{"z"}}, Eval{})
	var ue *UnboundError
	if !errors.As(err, &ue) || ue.Name != "z" {
		t.Fatalf("неизвестная переменная: %v; ожидали *UnboundError{z}", err)
	}
	if _, err := Walk[float64](Bin{'/', Num{1}, Bin{'-', Num{2}, Num{2}}}, Eval{}); !errors.Is(err, ErrDivByZero) {
		t.Fatalf("1 / (2 - 2): %v, ожидали ErrDivByZero", err)
	}
	if _, err := Walk[float64](Bin{'%', Num{1}, Num{2}}, Eval{}); err == nil {
		t.Fatal("неизвестная операция '%' должна давать ошибку")
	}
}

func TestVisitorNil(t *testing.T) {
	var np *Num
	cases := map[string]Node{"nil": nil, "nil-потомок": Bin{'+', Num{1}, nil}, "nil *Num": np, "nil *Bin": (*Bin)(nil)}
	for name, n := range cases {
		if _, err := Walk[string](n, Printer{}); !errors.Is(err, ErrNilNode) {
			t.Fatalf("%s: Walk = %v, ожидали ErrNilNode без паники", name, err)
		}
	}
	if _, err := Walk[string](chkAlien{}, Printer{}); err == nil {
		t.Fatal("неизвестный тип узла должен давать ошибку")
	}
}

func TestVisitorStopsOnError(t *testing.T) {
	var seen []string
	_, err := Walk[int](Bin{'+', Bin{'*', Var{"a"}, Var{"bad"}}, Var{"after"}}, chkCount{&seen})
	if err == nil || strings.Join(seen, ",") != "a,bad" {
		t.Fatalf("посещены %v, ошибка %v; ожидали обход слева направо и остановку на bad", seen, err)
	}
}

func TestVisitorFreeVars(t *testing.T) {
	e := Bin{'+', Bin{'*', Var{"b"}, Var{"a"}}, Bin{'-', Var{"b"}, Bin{'/', Var{"c"}, Num{1}}}}
	vars, err := Walk[[]string](e, FreeVars{})
	if err != nil || strings.Join(vars, ",") != "a,b,c" {
		t.Fatalf("FreeVars = %v, %v; ожидали [a b c]", vars, err)
	}
	if vars, _ := Walk[[]string](Num{1}, FreeVars{}); len(vars) != 0 {
		t.Fatalf("FreeVars(1) = %v, ожидали пусто", vars)
	}
}
