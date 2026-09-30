package main

// chkOpErr похож на net.OpError: у него есть Timeout(), но сам он таймаутом
// себя не считает — ответ лежит в причине.
type chkOpErr struct {
	op  string
	err error
}

func (e *chkOpErr) Error() string { return e.op + ": " + e.err.Error() }
func (e *chkOpErr) Unwrap() error { return e.err }
func (e *chkOpErr) Timeout() bool { return false }

type chkTimeout struct{}

func (chkTimeout) Error() string { return "i/o timeout" }
func (chkTimeout) Timeout() bool { return true }

func TestIsTimeoutBasic(t *testing.T) {
	if IsTimeout(nil) || IsTimeout(io.EOF) {
		t.Fatal("nil и io.EOF — не таймауты")
	}
	if !IsTimeout(fmt.Errorf("a: %w", chkTimeout{})) {
		t.Fatal("обёрнутый таймаут не найден")
	}
	if !IsTimeout(fmt.Errorf("запрос: %w", context.DeadlineExceeded)) {
		t.Fatal("context.DeadlineExceeded — таймаут (у него есть Timeout() == true)")
	}
}

func TestIsTimeoutDeeper(t *testing.T) {
	err := fmt.Errorf("get: %w", &chkOpErr{"read", chkTimeout{}})
	if !IsTimeout(err) {
		t.Fatal("внешнее звено сказало Timeout() == false, но глубже есть таймаут — ожидали true")
	}
	if IsTimeout(&chkOpErr{"dial", io.EOF}) {
		t.Fatal("звено с Timeout() == false и без таймаута внутри — не таймаут")
	}
}

func TestIsTimeoutJoin(t *testing.T) {
	err := errors.Join(io.EOF, &chkOpErr{"write", errors.Join(io.ErrClosedPipe, fmt.Errorf("x: %w", chkTimeout{}))})
	if !IsTimeout(err) {
		t.Fatal("таймаут во второй ветке вложенного Join не найден")
	}
	two := fmt.Errorf("a: %w, b: %w", io.EOF, context.DeadlineExceeded)
	if !IsTimeout(two) {
		t.Fatal("fmt.Errorf с двумя %w — таймаут во втором аргументе не найден")
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want Kind
	}{
		{"nil", nil, KindOK},
		{"дедлайн", fmt.Errorf("q: %w", context.DeadlineExceeded), KindTimeout},
		{"отмена", fmt.Errorf("q: %w", context.Canceled), KindCanceled},
		{"таймаут и отмена", errors.Join(context.Canceled, &chkOpErr{"read", chkTimeout{}}), KindTimeout},
		{"обычная", io.ErrUnexpectedEOF, KindFailure},
		{"Timeout() false", &chkOpErr{"dial", io.EOF}, KindFailure},
	}
	for _, c := range cases {
		if got := Classify(c.err); got != c.want {
			t.Fatalf("%s: Classify = %d, ожидали %d", c.name, got, c.want)
		}
	}
}
