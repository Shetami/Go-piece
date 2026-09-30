package main

func chkAcq(s *Weighted, n int64) <-chan error {
	ch := make(chan error, 1)
	go func() { ch <- s.Acquire(context.Background(), n) }()
	return ch
}

func chkGot(t *testing.T, ch <-chan error, what string) {
	t.Helper()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("%s: %v", what, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: не дождались", what)
	}
}

func chkStill(t *testing.T, ch <-chan error, what string) {
	t.Helper()
	select {
	case err := <-ch:
		t.Fatalf("%s: Acquire вернулся (%v), а должен ждать", what, err)
	case <-time.After(30 * time.Millisecond):
	}
}

func TestWeightedFIFO(t *testing.T) {
	s := NewWeighted(10)
	if err := s.Acquire(context.Background(), 8); err != nil {
		t.Fatal(err)
	}
	big := chkAcq(s, 5)
	chkStill(t, big, "занято 8 из 10, просим 5")
	if s.TryAcquire(2) {
		t.Fatal("TryAcquire(2) прошёл в обход ждущего Acquire(5) — большой запрос будет голодать")
	}
	small := chkAcq(s, 1)
	chkStill(t, small, "Acquire(1) стоит в очереди за Acquire(5)")
	s.Release(8)
	chkGot(t, big, "после Release(8) Acquire(5)")
	chkGot(t, small, "после Release(8) Acquire(1)")
	if !s.TryAcquire(4) || s.TryAcquire(1) {
		t.Fatal("занято 6 из 10: TryAcquire(4) должен пройти, следующий TryAcquire(1) — нет")
	}
}

func TestWeightedCancelFrontUnblocks(t *testing.T) {
	s := NewWeighted(10)
	s.Acquire(context.Background(), 5)
	ctx, cancel := context.WithCancel(context.Background())
	bigErr := make(chan error, 1)
	go func() { bigErr <- s.Acquire(ctx, 10) }()
	time.Sleep(20 * time.Millisecond)
	small := chkAcq(s, 3)
	chkStill(t, small, "Acquire(3) за ждущим Acquire(10)")
	cancel()
	select {
	case err := <-bigErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("отменённый Acquire вернул %v, ожидали context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("отмена не прервала Acquire")
	}
	chkGot(t, small, "первый в очереди отменён, свободно 5 — Acquire(3)")
	s.Release(8)
	if !s.TryAcquire(10) {
		t.Fatal("после Release всего занятого TryAcquire(10) не прошёл — отменённый запрос что-то занял?")
	}
}

func TestWeightedErrors(t *testing.T) {
	s := NewWeighted(4)
	if err := s.Acquire(context.Background(), 5); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Acquire(5) при ёмкости 4 = %v, ожидали ErrTooLarge сразу", err)
	}
	s.Acquire(context.Background(), 1)
	defer func() {
		if recover() == nil {
			t.Fatal("Release(2) при занятой 1 должен паниковать")
		}
	}()
	s.Release(2)
}

func TestWeightedStress(t *testing.T) {
	s := NewWeighted(6)
	var used, peak atomic.Int64
	var wg sync.WaitGroup
	for g := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				n := int64(1 + (g+i)%4)
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration((g+i)%3)*time.Millisecond)
				err := s.Acquire(ctx, n)
				cancel()
				if err != nil {
					continue
				}
				u := used.Add(n)
				for p := peak.Load(); u > p && !peak.CompareAndSwap(p, u); p = peak.Load() {
				}
				runtime.Gosched()
				used.Add(-n)
				s.Release(n)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("10 горутин с таймаутами зависли на семафоре")
	}
	if p := peak.Load(); p > 6 {
		t.Fatalf("одновременно занято %d единиц при ёмкости 6", p)
	}
	if !s.TryAcquire(6) {
		t.Fatal("после нагрузки семафор не вернулся к полной ёмкости — единицы потеряны")
	}
}
