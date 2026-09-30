package main

type chkClk struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClk) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkClk) Add(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

func TestCachedTTL(t *testing.T) {
	clk := &chkClk{t: time.Unix(1_700_000_000, 0)}
	calls := 0
	get := Cached(func(k string) (int, error) { calls++; return 0, nil }, time.Minute, clk.Now)
	get("a")
	clk.Add(59 * time.Second)
	get("a")
	if calls != 1 {
		t.Fatalf("через 59 с при ttl 1 мин load вызвана %d раз, ожидали 1 (и нулевой результат тоже кэшируется)", calls)
	}
	clk.Add(time.Second)
	get("a")
	if calls != 2 {
		t.Fatalf("ровно через ttl значение устарело: вызовов load %d, ожидали 2", calls)
	}
}

func TestCachedDedupAndErrorsShared(t *testing.T) {
	clk := &chkClk{t: time.Unix(1_700_000_000, 0)}
	var calls atomic.Int32
	release := make(chan struct{})
	fail := errors.New("upstream 503")
	get := Cached(func(k string) (string, error) {
		n := calls.Add(1)
		<-release
		if n == 1 {
			return "", fail
		}
		return "ok", nil
	}, time.Minute, clk.Now)
	var wg sync.WaitGroup
	var gotErr atomic.Int32
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := get("k"); errors.Is(err, fail) {
				gotErr.Add(1)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("20 одновременных вызовов с одним ключом: load вызвана %d раз, ожидали 1", n)
	}
	if n := gotErr.Load(); n != 20 {
		t.Fatalf("ошибку общей загрузки получили %d из 20 вызовов, ожидали все", n)
	}
	if v, err := get("k"); err != nil || v != "ok" {
		t.Fatalf("после ошибки следующий вызов должен грузить заново: %q, %v", v, err)
	}
}

func TestCachedKeysInParallel(t *testing.T) {
	clk := &chkClk{t: time.Unix(1_700_000_000, 0)}
	bStarted := make(chan struct{})
	get := Cached(func(k string) (string, error) {
		if k == "a" {
			select {
			case <-bStarted:
			case <-time.After(2 * time.Second):
				return "", errors.New("b так и не начал грузиться")
			}
		} else {
			close(bStarted)
		}
		return k, nil
	}, time.Minute, clk.Now)
	errc := make(chan error, 1)
	go func() { _, err := get("a"); errc <- err }()
	time.Sleep(20 * time.Millisecond)
	done := make(chan struct{})
	go func() { get("b"); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("загрузка b ждёт, пока закончится загрузка a — load вызывается под общей блокировкой")
	}
	if err := <-errc; err != nil {
		t.Fatalf("загрузка a: %v", err)
	}
}
