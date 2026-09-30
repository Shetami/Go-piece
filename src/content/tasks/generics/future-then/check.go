package main

func chkCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func TestFutureManyAwaiters(t *testing.T) {
	var calls atomic.Int32
	fu := Async(func() (int, error) {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})
	ctx, cancel := chkCtx()
	defer cancel()
	var wg sync.WaitGroup
	var bad atomic.Int32
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := fu.Await(ctx); v != 42 || err != nil {
				bad.Add(1)
			}
		}()
	}
	wg.Wait()
	if v, err := fu.Await(ctx); v != 42 || err != nil || bad.Load() != 0 {
		t.Fatalf("10 параллельных Await и ещё один: %d из них не получили 42, последний — (%v, %v)", bad.Load(), v, err)
	}
	if calls.Load() != 1 {
		t.Fatalf("f вызвана %d раз, ожидали ровно один", calls.Load())
	}
	select {
	case <-fu.Done():
	default:
		t.Fatalf("после результата канал Done не закрыт")
	}
}

func TestFutureAwaitCancelKeepsTask(t *testing.T) {
	release := make(chan struct{})
	fu := Async(func() (string, error) { <-release; return "готово", nil })
	short, cancelShort := context.WithCancel(context.Background())
	cancelShort()
	if v, err := fu.Await(short); !errors.Is(err, context.Canceled) || v != "" {
		t.Fatalf("Await с отменённым контекстом = (%q, %v), ожидали (\"\", context.Canceled)", v, err)
	}
	close(release)
	ctx, cancel := chkCtx()
	defer cancel()
	if v, err := fu.Await(ctx); v != "готово" || err != nil {
		t.Fatalf("после отменённого ожидания задача должна досчитаться: (%q, %v)", v, err)
	}
}

func TestFutureThenChain(t *testing.T) {
	release := make(chan struct{})
	src := Async(func() (int, error) { <-release; return 7, nil })
	start := time.Now()
	s := Then(src, func(x int) (string, error) { return strings.Repeat("*", x), nil })
	n := Then(s, func(s string) (int, error) { return len(s) * 2, nil })
	if time.Since(start) > time.Second {
		t.Fatalf("Then заблокировал вызывающего до готовности src")
	}
	close(release)
	ctx, cancel := chkCtx()
	defer cancel()
	if v, err := s.Await(ctx); v != "*******" || err != nil {
		t.Fatalf("Then(7, repeat) = (%q, %v)", v, err)
	}
	if v, err := n.Await(ctx); v != 14 || err != nil {
		t.Fatalf("Then(Then(…)) = (%v, %v), ожидали 14", v, err)
	}
}

func TestFutureThenError(t *testing.T) {
	boom := errors.New("нет связи")
	var gCalls atomic.Int32
	src := Async(func() (int, error) { return 0, boom })
	next := Then(src, func(x int) (float64, error) { gCalls.Add(1); return 1.5, nil })
	ctx, cancel := chkCtx()
	defer cancel()
	if v, err := next.Await(ctx); !errors.Is(err, boom) || v != 0 || gCalls.Load() != 0 {
		t.Fatalf("ошибка src: получили (%v, %v), g вызвана %d раз — ожидали (0, нет связи) и ни одного вызова", v, err, gCalls.Load())
	}
}

func TestFuturePanic(t *testing.T) {
	fu := Async(func() ([]int, error) {
		var m map[string]int
		m["x"] = 1
		return nil, nil
	})
	ctx, cancel := chkCtx()
	defer cancel()
	if _, err := fu.Await(ctx); err == nil || !strings.Contains(err.Error(), "panic") {
		t.Fatalf("паника в задаче: err = %v, ожидали ошибку со словом panic", err)
	}
}
