package main

func chkMaxConc(t *testing.T, procs int) {
	prev := runtime.GOMAXPROCS(procs)
	defer runtime.GOMAXPROCS(prev)
	var cur, peak atomic.Int32
	in := make([]int, 24)
	for i := range in {
		in[i] = i
	}
	out, err := ParallelMap(context.Background(), in, func(_ context.Context, v int) (int, error) {
		n := cur.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		cur.Add(-1)
		return v * v, nil
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка %v", err)
	}
	for i, v := range out {
		if v != i*i {
			t.Fatalf("out[%d] = %d, ожидали %d — порядок результатов нарушен", i, v, i*i)
		}
	}
	if got := runtime.GOMAXPROCS(0); got != procs {
		t.Fatalf("ParallelMap изменил GOMAXPROCS: было %d, стало %d", procs, got)
	}
	if p := peak.Load(); p != int32(procs) {
		t.Fatalf("при GOMAXPROCS=%d одновременно работало до %d вызовов f, ожидали ровно %d", procs, p, procs)
	}
}

func TestParallelMapLimit(t *testing.T) {
	chkMaxConc(t, 2)
	chkMaxConc(t, 3)
}

func TestParallelMapEmpty(t *testing.T) {
	out, err := ParallelMap(context.Background(), []string(nil), func(context.Context, string) (int, error) { return 1, nil })
	if err != nil || len(out) != 0 {
		t.Fatalf("пустой вход: %v, %v", out, err)
	}
}

func TestParallelMapFirstError(t *testing.T) {
	prev := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(prev)
	before := runtime.NumGoroutine()
	boom := errors.New("boom")
	var calls atomic.Int32
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	done := make(chan error, 1)
	go func() {
		_, err := ParallelMap(context.Background(), in, func(ctx context.Context, v int) (int, error) {
			calls.Add(1)
			if v == 0 {
				time.Sleep(20 * time.Millisecond) // остальные к этому моменту уже внутри f
				return 0, boom
			}
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(5 * time.Second):
				return v, nil
			}
		})
		done <- err
	}()
	var err error
	select {
	case err = <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("ParallelMap не вернулся за 3 с — контекст работающих f не отменён после ошибки")
	}
	if err != boom {
		t.Fatalf("вернули %v, ожидали первую ошибку boom (а не ошибку отменённых вызовов)", err)
	}
	if c := calls.Load(); c > 10 {
		t.Fatalf("после ошибки f вызвана %d раз из 100 — новые элементы запускать не надо", c)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > before; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после возврата осталось %d лишних горутин", n-before)
	}
}

func TestParallelMapParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ParallelMap(ctx, []int{1, 2, 3}, func(ctx context.Context, v int) (int, error) { return v, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый родительский контекст: err = %v, ожидали context.Canceled", err)
	}
}
