package main

func chkDo(t *testing.T, l *Limiter, ctx context.Context, fn func(context.Context) error) (err error, rec any) {
	t.Helper()
	type out struct {
		err error
		rec any
	}
	ch := make(chan out, 1)
	go func() {
		var o out
		defer func() { o.rec = recover(); ch <- o }()
		o.err = l.Do(ctx, fn)
	}()
	select {
	case o := <-ch:
		return o.err, o.rec
	case <-time.After(2 * time.Second):
		t.Fatal("Do завис — слоты утекли?")
		return nil, nil
	}
}

func TestLimiterLimit(t *testing.T) {
	l := NewLimiter(3)
	var cur, peak atomic.Int32
	var wg sync.WaitGroup
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Do(context.Background(), func(context.Context) error {
				n := cur.Add(1)
				for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
				}
				time.Sleep(2 * time.Millisecond)
				cur.Add(-1)
				return nil
			})
		}()
	}
	wg.Wait()
	if p := peak.Load(); p > 3 || p < 2 {
		t.Fatalf("одновременно выполнялось %d функций, ожидали от 2 до 3 при лимите 3", p)
	}
}

func TestLimiterReleasesOnPanicAndError(t *testing.T) {
	l := NewLimiter(2)
	bad := errors.New("503")
	for i := range 5 {
		if _, rec := chkDo(t, l, context.Background(), func(context.Context) error { panic(fmt.Sprint("сбой ", i)) }); rec != fmt.Sprint("сбой ", i) {
			t.Fatalf("паника fn должна лететь дальше, recover() = %v", rec)
		}
		if err, _ := chkDo(t, l, context.Background(), func(context.Context) error { return bad }); err != bad {
			t.Fatalf("Do = %v, ожидали ошибку fn", err)
		}
	}
	ran := 0
	for range 2 {
		chkDo(t, l, context.Background(), func(context.Context) error { ran++; return nil })
	}
	if ran != 2 {
		t.Fatalf("после паник и ошибок выполнено %d из 2 — слоты не вернулись", ran)
	}
}

func TestLimiterCancelWhileWaiting(t *testing.T) {
	l := NewLimiter(1)
	hold := make(chan struct{})
	started := make(chan struct{})
	go l.Do(context.Background(), func(context.Context) error { close(started); <-hold; return nil })
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	called := false
	err, _ := chkDo(t, l, ctx, func(context.Context) error { called = true; return nil })
	if !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatalf("ждали слот до дедлайна: Do = %v, fn вызвана = %v; ожидали DeadlineExceeded и false", err, called)
	}
	close(hold)
	if err, _ := chkDo(t, l, context.Background(), func(context.Context) error { return nil }); err != nil {
		t.Fatalf("после отмены ожидания слот должен быть свободен: %v", err)
	}
}

func TestLimiterCanceledCtxFreeSlot(t *testing.T) {
	l := NewLimiter(5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	for range 50 {
		l.Do(ctx, func(context.Context) error { calls++; return nil })
	}
	if calls != 0 {
		t.Fatalf("с отменённым ctx fn вызвана %d раз из 50, ожидали 0", calls)
	}
}
