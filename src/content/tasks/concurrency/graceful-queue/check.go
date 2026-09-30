package main

func TestQueueDrainsAcceptedTasks(t *testing.T) {
	base := runtime.NumGoroutine()
	q := NewQueue(2, 20)
	var done atomic.Int32
	for range 20 {
		if err := q.Submit(context.Background(), func() {
			time.Sleep(5 * time.Millisecond)
			done.Add(1)
		}); err != nil {
			t.Fatalf("Submit = %v", err)
		}
	}
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown = %v", err)
	}
	if n := done.Load(); n != 20 {
		t.Fatalf("Shutdown вернулся, выполнив %d задач из 20 принятых", n)
	}
	if err := q.Submit(context.Background(), func() {}); !errors.Is(err, ErrClosed) {
		t.Fatalf("Submit после Shutdown = %v, ожидали ErrClosed", err)
	}
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("повторный Shutdown = %v, ожидали nil", err)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine() - base; n > 0 {
		t.Fatalf("после Shutdown осталось %d лишних горутин", n)
	}
}

func TestQueueSubmitWaitsAndRespectsCtx(t *testing.T) {
	q := NewQueue(1, 1)
	release := make(chan struct{})
	q.Submit(context.Background(), func() { <-release }) // занял воркера
	time.Sleep(10 * time.Millisecond)
	q.Submit(context.Background(), func() {}) // занял буфер
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := q.Submit(ctx, func() {}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Submit в полный буфер с таймаутом = %v, ожидали DeadlineExceeded", err)
	}
	close(release)
	q.Shutdown(context.Background())
}

func TestQueueShutdownWhileSubmitBlocked(t *testing.T) {
	q := NewQueue(1, 1)
	release := make(chan struct{})
	var ran atomic.Int32
	q.Submit(context.Background(), func() { <-release; ran.Add(1) })
	time.Sleep(10 * time.Millisecond)
	q.Submit(context.Background(), func() { ran.Add(1) })
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for range 5 { // эти ждут места в буфере
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Submit запаниковал во время Shutdown: %v", r)
				}
			}()
			if q.Submit(context.Background(), func() { ran.Add(1) }) == nil {
				accepted.Add(1)
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	shut := make(chan error, 1)
	go func() { shut <- q.Shutdown(context.Background()) }()
	time.Sleep(20 * time.Millisecond)
	close(release)
	select {
	case err := <-shut:
		if err != nil {
			t.Fatalf("Shutdown = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown завис, пока Submit ждали места в буфере")
	}
	wg.Wait()
	if want := 2 + accepted.Load(); ran.Load() != want {
		t.Fatalf("выполнено %d задач, а принято %d — принятые задачи потеряны", ran.Load(), want)
	}
}

func TestQueueShutdownDeadline(t *testing.T) {
	q := NewQueue(1, 1)
	release := make(chan struct{})
	q.Submit(context.Background(), func() { <-release })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := q.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown с долгой задачей и таймаутом = %v, ожидали DeadlineExceeded", err)
	}
	close(release)
	done := make(chan error, 1)
	go func() { done <- q.Shutdown(context.Background()) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("второй Shutdown = %v, ожидали nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("второй Shutdown не дождался завершения воркеров")
	}
}
