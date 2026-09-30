package main

func chkPoll[T any](t *testing.T, ctx context.Context, interval time.Duration, check func(context.Context) (T, error)) (T, error) {
	t.Helper()
	type res struct {
		v   T
		err error
	}
	ch := make(chan res, 1)
	go func() {
		v, err := PollUntil(ctx, interval, check)
		ch <- res{v, err}
	}()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-time.After(2 * time.Second):
		t.Fatal("PollUntil не вернулся за 2 секунды")
	}
	var zero T
	return zero, nil
}

func TestPollFirstCallImmediate(t *testing.T) {
	calls := 0
	v, err := chkPoll(t, context.Background(), time.Hour, func(context.Context) (string, error) {
		calls++
		return "ready", nil
	})
	if v != "ready" || err != nil || calls != 1 {
		t.Fatalf("получили %q, %v после %d вызовов; ожидали \"ready\", nil с первого же вызова (не ждать interval)", v, err, calls)
	}
}

func TestPollRetriesNotReady(t *testing.T) {
	calls := 0
	v, err := chkPoll(t, context.Background(), time.Millisecond, func(context.Context) (int, error) {
		calls++
		if calls < 4 {
			return 0, fmt.Errorf("job %d: %w", calls, ErrNotReady)
		}
		return 42, nil
	})
	if v != 42 || err != nil || calls != 4 {
		t.Fatalf("получили %v, %v после %d вызовов; ожидали 42, nil после 4 (обёрнутая ErrNotReady — тоже «не готов»)", v, err, calls)
	}
}

func TestPollStopsOnFatal(t *testing.T) {
	fatal := errors.New("404")
	calls := 0
	_, err := chkPoll(t, context.Background(), time.Millisecond, func(context.Context) (int, error) {
		calls++
		if calls == 2 {
			return 0, fatal
		}
		return 0, ErrNotReady
	})
	if !errors.Is(err, fatal) || calls != 2 {
		t.Fatalf("ошибка %v после %d вызовов; ожидали 404 сразу на втором вызове", err, calls)
	}
}

func TestPollDeadlineKeepsCause(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	cause := fmt.Errorf("replica lag: %w", ErrNotReady)
	calls := 0
	_, err := chkPoll(t, ctx, 5*time.Millisecond, func(context.Context) (int, error) {
		calls++
		return 0, cause
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("по дедлайну получили %v, ожидали errors.Is(err, context.DeadlineExceeded)", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("ошибка %q потеряла последнюю причину «не готов»", err)
	}
	if calls < 2 {
		t.Fatalf("за 40 мс при interval 5 мс check вызван %d раз — опрос не повторялся", calls)
	}
}

func TestPollCancelledBefore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	_, err := chkPoll(t, ctx, time.Millisecond, func(context.Context) (int, error) {
		calls++
		return 1, nil
	})
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("ctx отменён заранее: %v, вызовов %d; ожидали context.Canceled и 0 вызовов", err, calls)
	}
}
