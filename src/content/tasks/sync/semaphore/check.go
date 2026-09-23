package main

func TestSemaphoreLimit(t *testing.T) {
	s := NewSemaphore(2)
	var cur, peak atomic.Int64
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if err := s.Acquire(context.Background()); err != nil {
				t.Errorf("Acquire: %v", err)
				return
			}
			defer s.Release()
			n := cur.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			cur.Add(-1)
		})
	}
	wg.Wait()
	if peak.Load() != 2 {
		t.Fatalf("одновременных владельцев было %d, ожидали ровно 2", peak.Load())
	}
}

func TestSemaphoreAcquireCanceled(t *testing.T) {
	s := NewSemaphore(1)
	if err := s.Acquire(context.Background()); err != nil {
		t.Fatalf("первый Acquire: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := s.Acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("места нет, контекст истёк — ожидали DeadlineExceeded, получили %v", err)
	}
	s.Release()
	if err := s.Acquire(context.Background()); err != nil {
		t.Fatalf("после Release место должно освободиться: %v", err)
	}
}
