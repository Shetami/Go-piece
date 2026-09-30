package main

var (
	chkA = errors.New("a")
	chkB = errors.New("b")
	chkC = errors.New("c")
	chkD = errors.New("d")
)

func chkErrs(err error) []error {
	if m, ok := err.(*MultiError); ok && m != nil {
		return m.Errs
	}
	return nil
}

func TestAppendNil(t *testing.T) {
	if err := Append(nil); err != nil {
		t.Fatalf("Append(nil) = %#v, ожидали настоящий nil", err)
	}
	if err := Append(nil, nil, nil); err != nil {
		t.Fatalf("Append(nil, nil, nil) = %#v, ожидали настоящий nil", err)
	}
	var m *MultiError
	if err := Append(m, nil); err != nil {
		t.Fatalf("Append((*MultiError)(nil), nil) = %#v, ожидали nil", err)
	}
}

func TestAppendSingle(t *testing.T) {
	if err := Append(nil, nil, chkA); err != chkA {
		t.Fatalf("одна ошибка возвращается как есть, получили %#v", err)
	}
}

func TestAppendFlat(t *testing.T) {
	err := Append(Append(chkA, chkB), nil, Append(chkC, chkD))
	m, ok := err.(*MultiError)
	if !ok || len(m.Errs) != 4 {
		t.Fatalf("ожидали плоский *MultiError из 4 ошибок, получили %#v", err)
	}
	if err.Error() != "ошибок: 4: a; b; c; d" {
		t.Fatalf("текст %q, ожидали %q", err.Error(), "ошибок: 4: a; b; c; d")
	}
	for _, e := range []error{chkA, chkB, chkC, chkD} {
		if !errors.Is(err, e) {
			t.Fatalf("errors.Is не находит %v", e)
		}
	}
	j := errors.Join(chkA, chkB)
	if err := Append(j, chkC); len(chkErrs(err)) != 2 {
		t.Fatalf("errors.Join — не наш тип, его не раскрываем: ожидали 2 элемента, получили %v", err)
	}
}

func TestAppendNoAliasing(t *testing.T) {
	m := &MultiError{Errs: make([]error, 2, 10)}
	m.Errs[0], m.Errs[1] = chkA, chkB
	x := Append(m, chkC)
	y := Append(m, chkD)
	xs, ys := chkErrs(x), chkErrs(y)
	if len(xs) != 3 || len(ys) != 3 || xs[2] != chkC || ys[2] != chkD {
		t.Fatalf("Append(m, c) и Append(m, d) испортили друг друга: %v и %v", xs, ys)
	}
	if len(m.Errs) != 2 {
		t.Fatalf("исходный MultiError изменился: %v", m.Errs)
	}
}
