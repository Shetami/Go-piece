package main

func chkShutdown(t *testing.T, s *Server, ctx context.Context) error {
	t.Helper()
	res := make(chan error, 1)
	go func() { res <- s.Shutdown(ctx) }()
	select {
	case err := <-res:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown не вернулся за 2 секунды")
	}
	return nil
}

func TestServerDrainsQueue(t *testing.T) {
	var handled atomic.Int64
	s := NewServer(2, 10, func(int) {
		time.Sleep(time.Millisecond)
		handled.Add(1)
	})
	for i := range 10 {
		if err := s.Submit(i); err != nil {
			t.Fatalf("Submit(%d) = %v", i, err)
		}
	}
	if err := chkShutdown(t, s, context.Background()); err != nil {
		t.Fatalf("Shutdown = %v", err)
	}
	if n := handled.Load(); n != 10 {
		t.Fatalf("после Shutdown обработано %d из 10 принятых задач — очередь не дочитана", n)
	}
}

func TestServerOverloaded(t *testing.T) {
	gate := make(chan struct{})
	started := make(chan int, 1)
	s := NewServer(1, 2, func(j int) {
		started <- j
		<-gate
	})
	s.Submit(1)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("воркер не взял первую задачу")
	}
	if err := s.Submit(2); err != nil {
		t.Fatalf("Submit(2) = %v", err)
	}
	if err := s.Submit(3); err != nil {
		t.Fatalf("Submit(3) = %v", err)
	}
	res := make(chan error, 1)
	go func() { res <- s.Submit(4) }()
	select {
	case err := <-res:
		if !errors.Is(err, ErrOverloaded) {
			t.Fatalf("Submit в полную очередь = %v, ожидали ErrOverloaded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Submit в полную очередь заблокировался")
	}
	close(gate)
	go func() {
		for range started {
		}
	}()
	chkShutdown(t, s, context.Background())
}

func TestServerShutdownTimeout(t *testing.T) {
	gate := make(chan struct{})
	var handled atomic.Int64
	s := NewServer(1, 5, func(int) { <-gate; handled.Add(1) })
	s.Submit(1)
	s.Submit(2)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := chkShutdown(t, s, ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown с зависшей задачей = %v, ожидали DeadlineExceeded", err)
	}
	if err := s.Submit(3); !errors.Is(err, ErrShuttingDown) {
		t.Fatalf("Submit после Shutdown = %v, ожидали ErrShuttingDown", err)
	}
	close(gate)
	if err := chkShutdown(t, s, context.Background()); err != nil {
		t.Fatalf("повторный Shutdown = %v, ожидали nil", err)
	}
	if handled.Load() != 2 {
		t.Fatalf("обработано %d, ожидали 2 — принятые задачи доделываются и после таймаута Shutdown", handled.Load())
	}
}

func TestServerSubmitRacesShutdown(t *testing.T) {
	var handled, accepted atomic.Int64
	s := NewServer(4, 8, func(int) { handled.Add(1) })
	panics := make(chan any, 20)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics <- r
				}
			}()
			for i := range 20000 { // предел — чтобы неверное решение не зависло навсегда
				err := s.Submit(i)
				if errors.Is(err, ErrShuttingDown) {
					return
				}
				if err == nil {
					accepted.Add(1)
				}
			}
		}()
	}
	// ждём по числу принятых, а не time.Sleep: пока отправители крутятся,
	// фиктивные часы песочницы не идут
	for i := 0; accepted.Load() < 200 && i < 1_000_000; i++ {
		runtime.Gosched()
	}
	if err := chkShutdown(t, s, context.Background()); err != nil {
		t.Fatalf("Shutdown = %v", err)
	}
	wg.Wait()
	select {
	case r := <-panics:
		t.Fatalf("Submit запаниковал во время Shutdown: %v", r)
	default:
	}
	if accepted.Load() != handled.Load() {
		t.Fatalf("принято %d задач, обработано %d — принятая задача потерялась", accepted.Load(), handled.Load())
	}
}
