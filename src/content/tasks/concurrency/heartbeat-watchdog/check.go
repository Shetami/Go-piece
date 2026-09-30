package main

func chkRunSupervise(t *testing.T, ctx context.Context, timeout time.Duration,
	work func(ctx context.Context, beat func()) error) error {
	t.Helper()
	res := make(chan error, 1)
	go func() { res <- Supervise(ctx, timeout, work) }()
	select {
	case err := <-res:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("Supervise не вернулся")
		return nil
	}
}

func TestSuperviseBeatingWorkerNotKilled(t *testing.T) {
	err := chkRunSupervise(t, context.Background(), 100*time.Millisecond, func(ctx context.Context, beat func()) error {
		for range 15 { // всего 300 мс — дольше timeout, но с сердцебиением
			time.Sleep(20 * time.Millisecond)
			if ctx.Err() != nil {
				return errors.New("работу убили, хотя сердцебиение шло")
			}
			beat()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Supervise = %v, ожидали nil", err)
	}
}

func TestSuperviseStalled(t *testing.T) {
	var finished atomic.Bool
	err := chkRunSupervise(t, context.Background(), 50*time.Millisecond, func(ctx context.Context, beat func()) error {
		beat()
		time.Sleep(20 * time.Millisecond)
		beat()
		<-ctx.Done() // зависли
		time.Sleep(20 * time.Millisecond)
		finished.Store(true)
		return ctx.Err()
	})
	if !errors.Is(err, ErrStalled) {
		t.Fatalf("Supervise = %v, ожидали ErrStalled", err)
	}
	if !finished.Load() {
		t.Fatal("Supervise вернулся раньше, чем work завершилась")
	}
}

func TestSuperviseWorkError(t *testing.T) {
	boom := errors.New("диск полон")
	err := chkRunSupervise(t, context.Background(), time.Second, func(ctx context.Context, beat func()) error {
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("Supervise = %v, ожидали ошибку work", err)
	}
}

func TestSuperviseParentCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	err := chkRunSupervise(t, ctx, time.Hour, func(ctx context.Context, beat func()) error {
		<-ctx.Done()
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Supervise = %v, ожидали DeadlineExceeded от внешнего ctx", err)
	}
}

func TestSuperviseBeatNeverBlocks(t *testing.T) {
	base := runtime.NumGoroutine()
	var saved func()
	chkRunSupervise(t, context.Background(), time.Second, func(ctx context.Context, beat func()) error {
		saved = beat
		for range 100 {
			beat() // много сердцебиений подряд, без пауз
		}
		return nil
	})
	done := make(chan struct{})
	go func() {
		for range 10 {
			saved() // после возврата Supervise
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("beat заблокировался после возврата Supervise")
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine() - base; n > 0 {
		t.Fatalf("осталось %d лишних горутин", n)
	}
}
