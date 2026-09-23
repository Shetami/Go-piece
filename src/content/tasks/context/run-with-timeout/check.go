package main

func TestRunWithTimeoutFast(t *testing.T) {
	boom := errors.New("своя ошибка")
	if err := RunWithTimeout(context.Background(), time.Second, func(context.Context) error { return boom }); !errors.Is(err, boom) {
		t.Fatalf("быстрая f: ожидали её ошибку, получили %v", err)
	}
	if err := RunWithTimeout(context.Background(), time.Second, func(context.Context) error { return nil }); err != nil {
		t.Fatalf("быстрая f без ошибки: %v", err)
	}
}

func TestRunWithTimeoutGivesDeadlineToF(t *testing.T) {
	err := RunWithTimeout(context.Background(), 30*time.Millisecond, func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); !ok {
			return errors.New("у контекста f нет дедлайна")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunWithTimeoutDoesNotWaitStubbornF(t *testing.T) {
	start := time.Now()
	err := RunWithTimeout(context.Background(), 30*time.Millisecond, func(context.Context) error {
		time.Sleep(2 * time.Second) // контекст игнорирует
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ожидали DeadlineExceeded, получили %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("RunWithTimeout ждал f, хотя срок вышел")
	}
}

func TestRunWithTimeoutParentCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunWithTimeout(ctx, time.Hour, func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый родитель: ожидали Canceled, получили %v", err)
	}
}
