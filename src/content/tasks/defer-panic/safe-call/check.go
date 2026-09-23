package main

func TestSafeCallNoPanic(t *testing.T) {
	called := false
	if err := SafeCall(func() { called = true }); err != nil {
		t.Fatalf("без паники ожидали nil, получили %v", err)
	}
	if !called {
		t.Fatal("f не вызвана")
	}
}

func TestSafeCallString(t *testing.T) {
	err := SafeCall(func() { panic("всё пропало") })
	if err == nil || !strings.Contains(err.Error(), "всё пропало") {
		t.Fatalf("паника строкой: ожидали ошибку с текстом, получили %v", err)
	}
}

var errCheckSentinel = errors.New("особая ошибка")

func TestSafeCallErrorIsKept(t *testing.T) {
	err := SafeCall(func() { panic(errCheckSentinel) })
	if !errors.Is(err, errCheckSentinel) {
		t.Fatalf("паника ошибкой должна находиться через errors.Is, получили %v", err)
	}
}

func TestSafeCallRuntimeError(t *testing.T) {
	err := SafeCall(func() {
		var m map[string]int
		m["x"] = 1
	})
	var re runtime.Error
	if !errors.As(err, &re) {
		t.Fatalf("паника рантайма должна достаться как runtime.Error, получили %v", err)
	}
}

func TestSafeCallNonErrorValue(t *testing.T) {
	err := SafeCall(func() { panic(42) })
	if err == nil || !strings.Contains(err.Error(), "42") {
		t.Fatalf("паника числом: ожидали ошибку с 42, получили %v", err)
	}
}
