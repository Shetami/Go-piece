package main

var errBad = errors.New("плохое число")

func parse(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", errBad, s)
	}
	return n, nil
}

func TestMustOK(t *testing.T) {
	if got := Must(parse("42")); got != 42 {
		t.Fatalf("Must(parse(\"42\")) = %d", got)
	}
}

func TestCatchNoPanic(t *testing.T) {
	sum := 0
	err := Catch(func() {
		sum = Must(parse("1")) + Must(parse("2"))
	})
	if err != nil || sum != 3 {
		t.Fatalf("без ошибок: err=%v sum=%d", err, sum)
	}
}

func TestCatchMustError(t *testing.T) {
	reached := false
	err := Catch(func() {
		_ = Must(parse("1")) + Must(parse("x"))
		reached = true
	})
	if !errors.Is(err, errBad) {
		t.Fatalf("Catch вернул %v, ожидали ошибку, оборачивающую errBad", err)
	}
	if reached {
		t.Fatalf("код после неудачного Must выполняться не должен")
	}
}

func TestCatchForeignPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != "чужая" {
			t.Fatalf("чужая паника должна пролететь сквозь Catch, получили %v", r)
		}
	}()
	Catch(func() { panic("чужая") })
	t.Fatalf("Catch проглотил чужую панику")
}

func TestCatchForeignErrorPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("panic(ошибка) без Must — тоже чужая паника, Catch не должен её ловить")
		}
	}()
	Catch(func() { panic(errBad) })
}
