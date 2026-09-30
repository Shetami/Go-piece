package main

func chkResult(t *testing.T, ch <-chan error, what string) error {
	t.Helper()
	if ch == nil {
		t.Fatal("Watchdog вернул nil-канал")
	}
	select {
	case err, ok := <-ch:
		if !ok {
			t.Fatalf("%s: канал закрыт без значения", what)
		}
		if _, ok := <-ch; ok {
			t.Fatalf("%s: после результата пришло второе значение", what)
		}
		return err
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: результата нет за 2 секунды", what)
	}
	return nil
}

func TestWatchdogKicksKeepAlive(t *testing.T) {
	kicks := make(chan struct{})
	res := Watchdog(context.Background(), 60*time.Millisecond, kicks)
	for i := range 20 {
		time.Sleep(10 * time.Millisecond)
		select {
		case kicks <- struct{}{}:
		case err := <-res:
			t.Fatalf("kick %d каждые 10 мс при timeout 60 мс, а watchdog сработал: %v (таймер не перезаводится?)", i, err)
		}
	}
	if err := chkResult(t, res, "после прекращения kick"); !errors.Is(err, ErrIdle) {
		t.Fatalf("kicks замолчали, результат %v, ожидали ErrIdle", err)
	}
}

func TestWatchdogIdleFromStart(t *testing.T) {
	res := Watchdog(context.Background(), 20*time.Millisecond, make(chan struct{}))
	if err := chkResult(t, res, "без единого kick"); !errors.Is(err, ErrIdle) {
		t.Fatalf("результат %v, ожидали ErrIdle", err)
	}
}

func TestWatchdogStopAndCancel(t *testing.T) {
	kicks := make(chan struct{})
	res := Watchdog(context.Background(), time.Hour, kicks)
	kicks <- struct{}{}
	close(kicks)
	if err := chkResult(t, res, "после закрытия kicks"); err != nil {
		t.Fatalf("kicks закрыт, результат %v, ожидали nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	res = Watchdog(ctx, time.Hour, make(chan struct{}))
	cancel()
	if err := chkResult(t, res, "после отмены"); !errors.Is(err, context.Canceled) {
		t.Fatalf("после отмены результат %v, ожидали context.Canceled", err)
	}
}

func TestWatchdogNoLeakWithoutReader(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 10 {
		Watchdog(context.Background(), time.Millisecond, make(chan struct{}))
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("результат никто не прочитал, и %d горутин Watchdog остались висеть", n-before)
	}
}
