package main

var chkErrBase = errors.New("нет соединения")

func chkInner() error { return WithStack(chkErrBase) }

func chkOuter() error { return fmt.Errorf("загрузка: %w", chkInner()) }

func TestStackNil(t *testing.T) {
	if err := WithStack(nil); err != nil {
		t.Fatalf("WithStack(nil) = %#v, ожидали настоящий nil", err)
	}
	if StackOf(errors.New("x")) != nil {
		t.Fatal("у ошибки без стека StackOf должен вернуть nil")
	}
}

func TestStackFrames(t *testing.T) {
	err := chkOuter()
	fr := StackOf(err)
	if len(fr) < 2 {
		t.Fatalf("стек %v слишком короткий", fr)
	}
	if !strings.HasSuffix(fr[0], ".chkInner") || !strings.HasSuffix(fr[1], ".chkOuter") {
		t.Fatalf("первые кадры %v, ожидали …chkInner, затем …chkOuter", fr[:2])
	}
	for _, f := range fr {
		if strings.HasSuffix(f, ".WithStack") || strings.HasPrefix(f, "runtime.Callers") {
			t.Fatalf("в стеке не должно быть служебных кадров, а есть %q", f)
		}
	}
	if len(fr) > 32 {
		t.Fatalf("кадров %d, ожидали не больше 32", len(fr))
	}
}

func TestStackTextAndUnwrap(t *testing.T) {
	err := chkInner()
	if err.Error() != "нет соединения" {
		t.Fatalf("Error() = %q, текст должен остаться прежним", err.Error())
	}
	if !errors.Is(err, chkErrBase) {
		t.Fatal("errors.Is не видит исходную ошибку сквозь *StackError")
	}
	var se *StackError
	if !errors.As(fmt.Errorf("x: %w", err), &se) || se.Err != chkErrBase {
		t.Fatal("errors.As должен достать *StackError с исходной ошибкой")
	}
}

func TestStackKeepsDeepest(t *testing.T) {
	inner := chkOuter()
	err := WithStack(inner)
	if err != inner {
		t.Fatalf("ошибку со стеком оборачивать второй раз не нужно, а получили %#v", err)
	}
	if fr := StackOf(err); len(fr) == 0 || !strings.HasSuffix(fr[0], ".chkInner") {
		t.Fatalf("StackOf должен вернуть самый глубокий стек (из chkInner), а вернул %v", fr)
	}
	joined := WithStack(errors.Join(errors.New("кэш"), chkInner()))
	if _, ok := joined.(*StackError); ok {
		t.Fatal("стек есть в одной из веток Join — повторно снимать не нужно")
	}
}

func chkDeep(n int) error {
	if n == 0 {
		return WithStack(chkErrBase)
	}
	return chkDeep(n - 1)
}

func TestStackLimit(t *testing.T) {
	if fr := StackOf(chkDeep(100)); len(fr) != 32 {
		t.Fatalf("на глубокой рекурсии кадров %d, ожидали ровно 32", len(fr))
	}
}
