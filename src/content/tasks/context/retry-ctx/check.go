package main

var errTemp = errors.New("временная ошибка")

func TestRetrySucceeds(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), 5, time.Millisecond, func(context.Context) error {
		calls++
		if calls < 3 {
			return errTemp
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("err=%v calls=%d, ожидали nil и 3 вызова", err, calls)
	}
}

func TestRetryExhausted(t *testing.T) {
	calls := 0
	start := time.Now()
	err := Retry(context.Background(), 3, 100*time.Millisecond, func(context.Context) error {
		calls++
		return fmt.Errorf("попытка %d: %w", calls, errTemp)
	})
	if calls != 3 {
		t.Fatalf("вызовов %d, ожидали 3", calls)
	}
	if err == nil || !strings.Contains(err.Error(), "попытка 3") {
		t.Fatalf("ожидали последнюю ошибку f, получили %v", err)
	}
	if d := time.Since(start); d > 280*time.Millisecond {
		t.Fatalf("три попытки с паузой 100 мс заняли %v — лишняя пауза после последней?", d)
	}
}

func TestRetryCancelDuringDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := Retry(ctx, 10, time.Hour, func(context.Context) error {
		calls++
		return errTemp
	})
	if time.Since(start) > time.Second {
		t.Fatalf("отмена не прервала ожидание между попытками")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ожидали ошибку с context.Canceled, получили %v", err)
	}
	if calls != 1 {
		t.Fatalf("вызовов %d, ожидали 1", calls)
	}
}

func TestRetryAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := Retry(ctx, 3, 0, func(context.Context) error { called = true; return nil })
	if called {
		t.Fatalf("контекст отменён заранее — f вызываться не должна")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ожидали context.Canceled, получили %v", err)
	}
}

func TestRetryPassesCtx(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "id-7")
	Retry(ctx, 1, 0, func(c context.Context) error {
		if c.Value(key{}) != "id-7" {
			t.Fatalf("f должна получить контекст вызывающего")
		}
		return nil
	})
}
