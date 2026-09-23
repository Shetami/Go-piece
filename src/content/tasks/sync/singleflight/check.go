package main

func TestSingleflightDedup(t *testing.T) {
	var g Group
	var calls atomic.Int64
	release := make(chan struct{})
	fn := func() (int, error) {
		calls.Add(1)
		<-release
		return 7, nil
	}

	var wg sync.WaitGroup
	results := make([]int, 10)
	for i := range 10 {
		wg.Go(func() {
			v, _ := g.Do("user:1", fn)
			results[i] = v
		})
	}
	time.Sleep(50 * time.Millisecond) // все успели встать в очередь
	close(release)
	wg.Wait()

	if calls.Load() != 1 {
		t.Fatalf("fn вызвана %d раз для одного ключа, ожидали 1", calls.Load())
	}
	for i, v := range results {
		if v != 7 {
			t.Fatalf("вызов %d получил %d, ожидали 7", i, v)
		}
	}
}

func TestSingleflightNotACache(t *testing.T) {
	var g Group
	calls := 0
	fn := func() (int, error) { calls++; return calls, nil }
	g.Do("k", fn)
	v, _ := g.Do("k", fn)
	if calls != 2 || v != 2 {
		t.Fatalf("после завершения вызов должен выполниться снова: calls=%d v=%d", calls, v)
	}
}

func TestSingleflightKeysIndependent(t *testing.T) {
	var g Group
	block := make(chan struct{})
	go g.Do("медленный", func() (int, error) { <-block; return 0, nil })
	time.Sleep(20 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		g.Do("быстрый", func() (int, error) { return 1, nil })
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("вызов с другим ключом ждал медленный вызов — fn выполняется под общей блокировкой?")
	}
	close(block)
}

func TestSingleflightError(t *testing.T) {
	var g Group
	boom := errors.New("упало")
	if _, err := g.Do("e", func() (int, error) { return 0, boom }); !errors.Is(err, boom) {
		t.Fatalf("ошибка fn должна вернуться, получили %v", err)
	}
}
