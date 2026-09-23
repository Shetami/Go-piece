package main

func TestLazyRetriesAfterError(t *testing.T) {
	var l Lazy[string]
	calls := 0
	init := func() (string, error) {
		calls++
		if calls < 3 {
			return "", errors.New("база недоступна")
		}
		return "соединение", nil
	}
	for i := 1; i <= 2; i++ {
		if _, err := l.Get(init); err == nil {
			t.Fatalf("вызов %d: ожидали ошибку инициализации", i)
		}
	}
	v, err := l.Get(init)
	if err != nil || v != "соединение" {
		t.Fatalf("третья попытка: %q, %v", v, err)
	}
	l.Get(init)
	if calls != 3 {
		t.Fatalf("после успеха init не должна вызываться: вызовов %d, ожидали 3", calls)
	}
}

func TestLazyConcurrentSingleInit(t *testing.T) {
	var l Lazy[int]
	var calls, running, peak atomic.Int64
	init := func() (int, error) {
		calls.Add(1)
		n := running.Add(1)
		if n > peak.Load() {
			peak.Store(n)
		}
		time.Sleep(20 * time.Millisecond)
		running.Add(-1)
		return 42, nil
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if v, err := l.Get(init); v != 42 || err != nil {
				t.Errorf("Get = %d, %v", v, err)
			}
		})
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("init вызвана %d раз, ожидали 1", calls.Load())
	}
	if peak.Load() > 1 {
		t.Fatalf("init выполнялась параллельно: %d одновременно", peak.Load())
	}
}
