package main

func chkWithin(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: не завершилось за 2 с", what)
	}
}

func TestQueueDrainsOnShutdown(t *testing.T) {
	s := NewServer(2, 20)
	var n atomic.Int32
	for range 20 {
		if err := s.Submit(context.Background(), func(context.Context) { time.Sleep(3 * time.Millisecond); n.Add(1) }); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}
	var err error
	chkWithin(t, "Shutdown", func() { err = s.Shutdown(context.Background()) })
	if err != nil || n.Load() != 20 {
		t.Fatalf("Shutdown = %v, выполнено %d из 20; ожидали дочитать очередь", err, n.Load())
	}
	if err := s.Submit(context.Background(), func(context.Context) {}); !errors.Is(err, ErrClosed) {
		t.Fatalf("Submit после Shutdown = %v, ожидали ErrClosed", err)
	}
	chkWithin(t, "повторный Shutdown", func() { err = s.Shutdown(context.Background()) })
}

func TestQueueFullSubmit(t *testing.T) {
	s := NewServer(1, 1)
	gate := make(chan struct{})
	started := make(chan struct{})
	s.Submit(context.Background(), func(context.Context) { close(started); <-gate })
	<-started
	s.Submit(context.Background(), func(context.Context) {}) // занимает единственное место
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	var err error
	chkWithin(t, "Submit в полную очередь с таймаутом", func() { err = s.Submit(ctx, func(context.Context) {}) })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Submit в полную очередь = %v, ожидали DeadlineExceeded", err)
	}
	// Submit ждёт места без таймаута — Shutdown должен его разбудить.
	res := make(chan error, 1)
	go func() { res <- s.Submit(context.Background(), func(context.Context) {}) }()
	time.Sleep(20 * time.Millisecond)
	go s.Shutdown(context.Background())
	select {
	case err := <-res:
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("ждавший места Submit = %v, ожидали ErrClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Submit, ждавший места, не проснулся при Shutdown")
	}
	close(gate)
}

func TestQueueForcedShutdown(t *testing.T) {
	s := NewServer(1, 10)
	sawCancel := make(chan struct{})
	var later atomic.Int32
	s.Submit(context.Background(), func(ctx context.Context) { <-ctx.Done(); close(sawCancel) })
	for range 3 {
		s.Submit(context.Background(), func(context.Context) { later.Add(1) })
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	var err error
	chkWithin(t, "Shutdown с таймаутом", func() { err = s.Shutdown(ctx) })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown = %v, ожидали DeadlineExceeded", err)
	}
	select {
	case <-sawCancel:
	default:
		t.Fatalf("Shutdown вернулся, а зависшее задание не получило отмену или ещё работает")
	}
	if later.Load() != 0 {
		t.Fatalf("после принудительной остановки выполнено ещё %d заданий из очереди", later.Load())
	}
}

func TestQueueConcurrentSubmitShutdown(t *testing.T) {
	s := NewServer(4, 8)
	var accepted, ran atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				err := s.Submit(context.Background(), func(context.Context) { ran.Add(1) })
				if err == nil {
					accepted.Add(1)
				} else if !errors.Is(err, ErrClosed) {
					t.Errorf("Submit = %v, ожидали nil или ErrClosed", err)
					return
				}
			}
		}()
	}
	time.Sleep(time.Millisecond)
	chkWithin(t, "Shutdown под нагрузкой", func() { s.Shutdown(context.Background()) })
	chkWithin(t, "Submit после Shutdown", wg.Wait)
	if accepted.Load() != ran.Load() {
		t.Fatalf("принято %d заданий, выполнено %d — принятые задания потерялись", accepted.Load(), ran.Load())
	}
}
