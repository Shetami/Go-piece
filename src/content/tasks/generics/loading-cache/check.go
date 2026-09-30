package main

type chkClock struct{ ns atomic.Int64 }

func (c *chkClock) Now() time.Time          { return time.Unix(0, c.ns.Load()) }
func (c *chkClock) Add(d time.Duration)     { c.ns.Add(int64(d)) }

func chkWait(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: зависло", what)
	}
}

func TestCacheSingleLoad(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	c := NewCache(time.Minute, nil, func(ctx context.Context, id int) (string, error) {
		calls.Add(1)
		<-release
		return fmt.Sprintf("user-%d", id), nil
	})
	var wg sync.WaitGroup
	var bad atomic.Int32
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := c.Get(context.Background(), 7); v != "user-7" || err != nil {
				bad.Add(1)
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	chkWait(t, "50 Get одного ключа", wg.Wait)
	if calls.Load() != 1 || bad.Load() != 0 {
		t.Fatalf("50 одновременных Get: load вызван %d раз (ожидали 1), неверных ответов %d", calls.Load(), bad.Load())
	}
}

func TestCacheKeysIndependent(t *testing.T) {
	slow := make(chan struct{})
	defer close(slow)
	c := NewCache(time.Minute, nil, func(ctx context.Context, k string) (int, error) {
		if k == "slow" {
			<-slow
		}
		return len(k), nil
	})
	go c.Get(context.Background(), "slow")
	time.Sleep(10 * time.Millisecond)
	chkWait(t, "Get другого ключа во время медленной загрузки", func() {
		if v, err := c.Get(context.Background(), "fast"); v != 4 || err != nil {
			t.Errorf("Get(fast) = (%v, %v)", v, err)
		}
	})
}

func TestCacheTTL(t *testing.T) {
	var clk chkClock
	var calls atomic.Int32
	c := NewCache(10*time.Second, clk.Now, func(ctx context.Context, k string) (int32, error) {
		return calls.Add(1), nil
	})
	ctx := context.Background()
	c.Get(ctx, "a")
	clk.Add(10*time.Second - 1)
	if v, _ := c.Get(ctx, "a"); v != 1 {
		t.Fatalf("за наносекунду до истечения ttl запись ещё свежая, а получили загрузку №%d", v)
	}
	clk.Add(1)
	if v, _ := c.Get(ctx, "a"); v != 2 {
		t.Fatalf("ровно через ttl запись устарела — ожидали новую загрузку №2, получили %d", v)
	}
}

func TestCacheErrorsNotCached(t *testing.T) {
	fail := true
	c := NewCache(time.Minute, nil, func(ctx context.Context, k string) (string, error) {
		if fail {
			fail = false
			return "", errors.New("база недоступна")
		}
		return "ok", nil
	})
	if _, err := c.Get(context.Background(), "k"); err == nil {
		t.Fatalf("первая загрузка упала, а ошибки нет")
	}
	if v, err := c.Get(context.Background(), "k"); v != "ok" || err != nil {
		t.Fatalf("ошибка закэшировалась: второй Get = (%q, %v)", v, err)
	}
}

func TestCacheCancelWaiter(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	c := NewCache(time.Minute, nil, func(ctx context.Context, k string) (string, error) {
		calls.Add(1)
		<-release
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "профиль", nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	chkWait(t, "Get с отменяемым контекстом", func() {
		if _, err := c.Get(ctx, "u"); !errors.Is(err, context.Canceled) {
			t.Errorf("Get с отменённым контекстом вернул %v, ожидали context.Canceled", err)
		}
	})
	close(release)
	chkWait(t, "Get после отмены", func() {
		if v, err := c.Get(context.Background(), "u"); v != "профиль" || err != nil {
			t.Errorf("загрузка должна была дойти до конца без отмены: (%q, %v)", v, err)
		}
	})
	if calls.Load() != 1 {
		t.Errorf("load вызван %d раз, ожидали 1", calls.Load())
	}
}
