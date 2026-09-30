package main

func chkDispatch(t *testing.T, ctx context.Context, in <-chan Task, n int, handle func(Task)) error {
	t.Helper()
	res := make(chan error, 1)
	go func() { res <- Dispatch(ctx, in, n, handle) }()
	select {
	case err := <-res:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("Dispatch не вернулся за 3 секунды")
	}
	return nil
}

func TestDispatchOrderPerKey(t *testing.T) {
	in := make(chan Task)
	go func() {
		for seq := range 50 {
			for k := range 10 {
				in <- Task{Key: fmt.Sprintf("user-%d", k), Seq: seq}
			}
		}
		close(in)
	}()
	var mu sync.Mutex
	seen := map[string][]int{}
	var cur, peak atomic.Int64
	err := chkDispatch(t, context.Background(), in, 4, func(tk Task) {
		c := cur.Add(1)
		for p := peak.Load(); c > p && !peak.CompareAndSwap(p, c); p = peak.Load() {
		}
		time.Sleep(20 * time.Microsecond)
		mu.Lock()
		seen[tk.Key] = append(seen[tk.Key], tk.Seq)
		mu.Unlock()
		cur.Add(-1)
	})
	if err != nil {
		t.Fatalf("Dispatch = %v, ожидали nil", err)
	}
	if len(seen) != 10 {
		t.Fatalf("обработано ключей: %d, ожидали 10 (к моменту возврата должно быть обработано всё)", len(seen))
	}
	for k, seqs := range seen {
		if len(seqs) != 50 || !slices.IsSorted(seqs) {
			t.Fatalf("ключ %s: задачи %v — ожидали 50 задач строго по порядку", k, seqs)
		}
	}
	if p := peak.Load(); p > 4 {
		t.Fatalf("одновременно работало %d handle при n = 4", p)
	}
	if p := peak.Load(); p < 2 {
		t.Fatalf("разные ключи обрабатывались строго по одному (максимум %d одновременно)", p)
	}
}

func TestDispatchCancel(t *testing.T) {
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan Task) // не закрывается никогда
	go func() {
		in <- Task{Key: "a"}
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	var handled atomic.Int64
	err := chkDispatch(t, ctx, in, 3, func(Task) { handled.Add(1) })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("после отмены Dispatch = %v, ожидали context.Canceled", err)
	}
	if handled.Load() != 1 {
		t.Fatalf("обработано %d задач, ожидали 1", handled.Load())
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before+1 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before+1 {
		t.Fatalf("после отмены остались горутины воркеров: %d лишних", n-before)
	}
}

func TestDispatchWaitsRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan Task, 1)
	in <- Task{Key: "slow"}
	var finished atomic.Bool
	started := make(chan struct{})
	go func() { <-started; cancel() }()
	err := chkDispatch(t, ctx, in, 2, func(Task) {
		close(started)
		time.Sleep(30 * time.Millisecond)
		finished.Store(true)
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Dispatch = %v, ожидали context.Canceled", err)
	}
	if !finished.Load() {
		t.Fatal("Dispatch вернулся, не дождавшись начатого handle")
	}
}
