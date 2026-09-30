package main

func chkEvery(t *testing.T, ctx context.Context, d time.Duration, f func(context.Context) error) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- Every(ctx, d, f) }()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatalf("Every не вернулась за 3 с")
		return nil
	}
}

func TestEveryFirstCallImmediate(t *testing.T) {
	stop := errors.New("стоп")
	start := time.Now()
	err := chkEvery(t, context.Background(), time.Hour, func(context.Context) error { return stop })
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v, ожидали ошибку f", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("первый вызов ждал целый interval — он должен быть сразу")
	}
}

func TestEveryStopsOnError(t *testing.T) {
	bad := errors.New("база недоступна")
	calls := 0
	err := chkEvery(t, context.Background(), 5*time.Millisecond, func(context.Context) error {
		calls++
		if calls == 3 {
			return bad
		}
		return nil
	})
	if !errors.Is(err, bad) || calls != 3 {
		t.Fatalf("err = %v, calls = %d; ожидали остановку на третьем вызове с его ошибкой", err, calls)
	}
}

func TestEveryCancelDuringWait(t *testing.T) {
	why := errors.New("сервис останавливается")
	ctx, cancel := context.WithCancelCause(context.Background())
	var calls atomic.Int32
	go func() { time.Sleep(30 * time.Millisecond); cancel(why) }()
	err := chkEvery(t, ctx, time.Hour, func(context.Context) error { calls.Add(1); return nil })
	if !errors.Is(err, why) || calls.Load() != 1 {
		t.Fatalf("err = %v, calls = %d; ожидали один вызов и причину отмены", err, calls.Load())
	}
	time.Sleep(20 * time.Millisecond)
	if calls.Load() != 1 {
		t.Fatalf("f вызывается после возврата Every")
	}
}

func TestEveryNoOverlapNoBurst(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var cur, peak, calls atomic.Int32
	var gotCtx atomic.Bool
	go func() { time.Sleep(235 * time.Millisecond); cancel() }()
	chkEvery(t, ctx, 10*time.Millisecond, func(c context.Context) error {
		if c.Done() != nil {
			gotCtx.Store(true)
		}
		if cur.Add(1) > 1 {
			peak.Store(2)
		}
		if calls.Add(1) == 1 {
			time.Sleep(200 * time.Millisecond) // долгий первый вызов: 20 тиков пропущено
		}
		cur.Add(-1)
		return nil
	})
	if peak.Load() > 1 {
		t.Fatalf("вызовы f перекрылись — следующий начался до конца предыдущего")
	}
	if !gotCtx.Load() {
		t.Fatalf("f получила не тот контекст, который передали в Every")
	}
	// Ожидаем: 1 долгий + 1 вдогонку + ~3 по расписанию. Пачка пропущенных — это 20+.
	if n := calls.Load(); n < 2 || n > 8 {
		t.Fatalf("вызовов %d; после долгого вызова пропущенные тики не должны выстреливать пачкой", n)
	}
}
