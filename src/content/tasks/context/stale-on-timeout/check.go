package main

// chkClock — ручные часы для TTL.
type chkClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkClock) Add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

// chkSrc — источник: отдаёт "v<N>" за delay, считает вызовы.
type chkSrc struct {
	calls atomic.Int32
	delay atomic.Int64
	fail  atomic.Pointer[error]
	sawOK atomic.Bool // контекст загрузки был жив в момент ответа
}

func (s *chkSrc) load(ctx context.Context, k string) (string, error) {
	n := s.calls.Add(1)
	select {
	case <-time.After(time.Duration(s.delay.Load())):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	s.sawOK.Store(ctx.Err() == nil)
	if e := s.fail.Load(); e != nil {
		return "", *e
	}
	return fmt.Sprintf("%s-v%d", k, n), nil
}

func chkCache() (*Cache[string, string], *chkSrc, *chkClock) {
	src, clk := &chkSrc{}, &chkClock{t: time.Unix(1_700_000_000, 0)}
	return NewCache[string, string](time.Minute, time.Second, src.load, clk.Now), src, clk
}

func TestCacheFreshHit(t *testing.T) {
	c, src, clk := chkCache()
	v, stale, err := c.Get(context.Background(), "a")
	if v != "a-v1" || stale || err != nil {
		t.Fatalf("первый Get = %q, %v, %v; ожидали a-v1", v, stale, err)
	}
	clk.Add(30 * time.Second)
	if v, _, _ := c.Get(context.Background(), "a"); v != "a-v1" || src.calls.Load() != 1 {
		t.Fatalf("свежее значение: Get = %q, загрузок %d; ожидали из кэша без загрузки", v, src.calls.Load())
	}
	clk.Add(31 * time.Second)
	if v, stale, _ := c.Get(context.Background(), "a"); v != "a-v2" || stale {
		t.Fatalf("после TTL Get = %q (stale %v); ожидали новую загрузку a-v2", v, stale)
	}
}

func TestCacheStaleOnTimeoutAndBackgroundRefresh(t *testing.T) {
	c, src, clk := chkCache()
	c.Get(context.Background(), "a")
	clk.Add(2 * time.Minute)
	src.delay.Store(int64(100 * time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	v, stale, err := c.Get(ctx, "a")
	if v != "a-v1" || !stale || err != nil {
		t.Fatalf("медленная загрузка: Get = %q, %v, %v; ожидали устаревшее a-v1, stale = true", v, stale, err)
	}
	time.Sleep(200 * time.Millisecond)
	if !src.sawOK.Load() {
		t.Fatalf("загрузку отменили вместе с вызывающим — она должна жить своим таймаутом")
	}
	src.delay.Store(0)
	if v, stale, _ := c.Get(context.Background(), "a"); v != "a-v2" || stale || src.calls.Load() != 2 {
		t.Fatalf("после фоновой загрузки Get = %q (stale %v, загрузок %d); ожидали a-v2 из кэша", v, stale, src.calls.Load())
	}
}

func TestCacheNoStaleTimeout(t *testing.T) {
	c, src, _ := chkCache()
	src.delay.Store(int64(time.Hour))
	why := errors.New("дедлайн запроса")
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel(why) }()
	_, _, err := c.Get(ctx, "b")
	if !errors.Is(err, why) {
		t.Fatalf("устаревшего нет, вызывающий ушёл: err = %v, ожидали причину отмены", err)
	}
}

func TestCacheLoadErrors(t *testing.T) {
	c, src, clk := chkCache()
	down := errors.New("база недоступна")
	src.fail.Store(&down)
	if _, _, err := c.Get(context.Background(), "x"); !errors.Is(err, down) {
		t.Fatalf("ошибка загрузки без устаревшего: err = %v", err)
	}
	src.fail.Store(nil)
	if v, _, err := c.Get(context.Background(), "x"); err != nil || v != "x-v2" {
		t.Fatalf("ошибка закэширована: Get = %q, %v", v, err)
	}
	clk.Add(2 * time.Minute)
	src.fail.Store(&down)
	if v, stale, err := c.Get(context.Background(), "x"); v != "x-v2" || !stale || err != nil {
		t.Fatalf("ошибка загрузки при устаревшем: Get = %q, %v, %v; ожидали x-v2, stale", v, stale, err)
	}
}

func TestCacheOneLoadPerKey(t *testing.T) {
	c, src, _ := chkCache()
	src.delay.Store(int64(30 * time.Millisecond))
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() { defer wg.Done(); c.Get(context.Background(), "hot") }()
	}
	wg.Wait()
	if n := src.calls.Load(); n != 1 {
		t.Fatalf("20 одновременных промахов — %d загрузок, ожидали 1", n)
	}
}
