package main

func TestQueueFullAndFIFO(t *testing.T) {
	q := NewQueue[int](2)
	if err := q.TryPush(1); err != nil {
		t.Fatalf("TryPush(1) = %v, ожидали nil", err)
	}
	if err := q.TryPush(2); err != nil {
		t.Fatalf("TryPush(2) = %v, ожидали nil", err)
	}
	if err := q.TryPush(3); !errors.Is(err, ErrFull) {
		t.Fatalf("TryPush в полную очередь = %v, ожидали ErrFull", err)
	}
	if n := q.Len(); n != 2 {
		t.Fatalf("Len = %d, ожидали 2", n)
	}
	ctx := context.Background()
	if v, err := q.Pop(ctx); v != 1 || err != nil {
		t.Fatalf("Pop = %v, %v; ожидали 1, nil", v, err)
	}
	if err := q.TryPush(3); err != nil {
		t.Fatalf("после Pop место есть, а TryPush = %v", err)
	}
	for _, want := range []int{2, 3} {
		if v, _ := q.Pop(ctx); v != want {
			t.Fatalf("Pop = %d, ожидали %d (порядок FIFO)", v, want)
		}
	}
}

func TestQueuePopWaitsAndTimesOut(t *testing.T) {
	q := NewQueue[string](1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	v, err := q.Pop(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || v != "" {
		t.Fatalf("Pop из пустой очереди по таймауту = %q, %v; ожидали \"\", DeadlineExceeded", v, err)
	}
	go func() {
		time.Sleep(10 * time.Millisecond)
		q.TryPush("x")
	}()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	if v, err := q.Pop(ctx2); v != "x" || err != nil {
		t.Fatalf("Pop должен дождаться элемента: получили %q, %v", v, err)
	}
}

func TestQueuePopPrefersItem(t *testing.T) {
	q := NewQueue[int](1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := range 100 {
		q.TryPush(i)
		v, err := q.Pop(ctx)
		if err != nil || v != i {
			t.Fatalf("элемент лежит в очереди, а Pop с отменённым ctx = %v, %v", v, err)
		}
	}
}

func TestQueuePushWait(t *testing.T) {
	q := NewQueue[int](1)
	q.TryPush(1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := q.PushWait(ctx, 2); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("PushWait в полную очередь по таймауту = %v, ожидали DeadlineExceeded", err)
	}
	done := make(chan error, 1)
	go func() { done <- q.PushWait(context.Background(), 3) }()
	time.Sleep(10 * time.Millisecond)
	if v, _ := q.Pop(context.Background()); v != 1 {
		t.Fatalf("Pop = %d, ожидали 1", v)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("PushWait после освобождения места = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("PushWait не дождался освободившегося места")
	}
}

func TestQueueConcurrentPush(t *testing.T) {
	q := NewQueue[int](50)
	var ok atomic.Int64
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				if q.TryPush(g*100+i) == nil {
					ok.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 50 || q.Len() != 50 {
		t.Fatalf("принято %d, Len = %d; ожидали ровно 50 при ёмкости 50", ok.Load(), q.Len())
	}
}
