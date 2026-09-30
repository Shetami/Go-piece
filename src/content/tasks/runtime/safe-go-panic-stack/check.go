package main

func chkExplode() error {
	var m map[string]int
	m["boom"] = 1 // паника: запись в nil-карту
	return nil
}

func chkRecv(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatalf("результат не пришёл за 3 с")
		return nil
	}
}

func TestGoResult(t *testing.T) {
	if err := chkRecv(t, Go(func() error { return nil })); err != nil {
		t.Fatalf("f вернула nil, получили %v", err)
	}
	e := errors.New("обычная ошибка")
	if err := chkRecv(t, Go(func() error { return e })); err != e {
		t.Fatalf("f вернула ошибку, получили %v", err)
	}
}

func TestGoPanicStack(t *testing.T) {
	err := chkRecv(t, Go(chkExplode))
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("паника должна прийти как *PanicError, получили %T %v", err, err)
	}
	if !strings.HasPrefix(pe.Error(), "panic: ") || !strings.Contains(pe.Error(), "nil map") {
		t.Fatalf("Error() = %q, ожидали \"panic: …nil map…\"", pe.Error())
	}
	if !strings.Contains(string(pe.Stack), "chkExplode") {
		t.Fatalf("в Stack нет функции chkExplode — стек снят не в момент паники:\n%s", pe.Stack)
	}
}

func TestGoPanicUnwrap(t *testing.T) {
	err := chkRecv(t, Go(func() error { panic(io.ErrUnexpectedEOF) }))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("panic(io.ErrUnexpectedEOF): errors.Is не видит исходную ошибку в %v", err)
	}
	err = chkRecv(t, Go(func() error { panic(nil) }))
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("panic(nil) тоже паника: получили %v", err)
	}
}

func TestGoGoexit(t *testing.T) {
	err := chkRecv(t, Go(func() error { runtime.Goexit(); return nil }))
	if err != ErrGoexit {
		t.Fatalf("runtime.Goexit: получили %v, ожидали ErrGoexit", err)
	}
}

func TestGoNoReaderNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 50; i++ {
		Go(func() error { return nil })
		Go(func() error { panic("x") })
	}
	for i := 0; i < 200 && runtime.NumGoroutine() > before; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("результат никто не читает, и %d горутин повисли на отправке", n-before)
	}
}
