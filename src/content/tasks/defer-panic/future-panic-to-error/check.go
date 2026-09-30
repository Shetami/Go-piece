package main

func chkBadLoader() (int, error) {
	var m map[string]int
	m["k"] = 1
	return 1, nil
}

func chkGetWithin[T any](t *testing.T, f *Future[T], ctx context.Context) (T, error) {
	t.Helper()
	type out struct {
		v   T
		err error
	}
	ch := make(chan out, 1)
	go func() { v, err := f.Get(ctx); ch <- out{v, err} }()
	select {
	case o := <-ch:
		return o.v, o.err
	case <-time.After(3 * time.Second):
		t.Fatal("Get не вернулся за 3 секунды")
		var zero T
		return zero, nil
	}
}

func TestFutureValue(t *testing.T) {
	var calls atomic.Int32
	f := Async(func() (string, error) { calls.Add(1); time.Sleep(5 * time.Millisecond); return "профиль", nil })
	var wg sync.WaitGroup
	var bad atomic.Int32
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := f.Get(context.Background()); v != "профиль" || err != nil {
				bad.Add(1)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("10 одновременных Get зависли — результат получил только один?")
	}
	if bad.Load() != 0 || calls.Load() != 1 {
		t.Fatalf("неверных результатов %d, вызовов fn %d; ожидали 0 и 1", bad.Load(), calls.Load())
	}
	if v, _ := chkGetWithin(t, f, context.Background()); v != "профиль" {
		t.Fatalf("повторный Get = %q", v)
	}
}

func TestFutureError(t *testing.T) {
	notFound := errors.New("не найдено")
	f := Async(func() (int, error) { return 0, notFound })
	if _, err := chkGetWithin(t, f, context.Background()); err != notFound {
		t.Fatalf("Get = %v, ожидали ошибку fn", err)
	}
}

func TestFuturePanic(t *testing.T) {
	f := Async(chkBadLoader)
	for range 2 {
		v, err := chkGetWithin(t, f, context.Background())
		var pe *PanicError
		if !errors.As(err, &pe) || v != 0 {
			t.Fatalf("Get = %d, %v; ожидали 0 и *PanicError", v, err)
		}
		var re runtime.Error
		if !errors.As(err, &re) {
			t.Fatalf("паника рантайма должна находиться через errors.As: %v", err)
		}
		if !strings.Contains(string(pe.Stack), "chkBadLoader") {
			t.Fatal("в Stack нет функции chkBadLoader, где случилась паника")
		}
	}
}

func TestFutureCancelThenResult(t *testing.T) {
	release := make(chan struct{})
	f := Async(func() (int, error) { <-release; return 7, nil })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if v, err := chkGetWithin(t, f, ctx); !errors.Is(err, context.Canceled) || v != 0 {
		t.Fatalf("Get с отменённым ctx = %d, %v; ожидали 0, context.Canceled", v, err)
	}
	close(release)
	if v, err := chkGetWithin(t, f, context.Background()); v != 7 || err != nil {
		t.Fatalf("после отмены первого Get следующий = %d, %v; ожидали 7, nil", v, err)
	}
}
