package main

var chkErrDown = errors.New("источник недоступен")

type chkSrc struct {
	mu    sync.Mutex
	calls int
	val   string
	err   error
}

func (s *chkSrc) load(string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.val, s.err
}

func (s *chkSrc) set(v string, err error) { s.mu.Lock(); s.val, s.err = v, err; s.mu.Unlock() }

type chkNow struct{ t time.Time }

func (c *chkNow) now() time.Time { return c.t }

func TestCacheFresh(t *testing.T) {
	src := &chkSrc{val: "v1"}
	clk := &chkNow{time.Unix(0, 0)}
	c := NewCache(src.load, time.Minute, 10*time.Second, clk.now)
	for range 3 {
		if v, err := c.Get("k"); v != "v1" || err != nil {
			t.Fatalf("Get = %q, %v; ожидали v1, nil", v, err)
		}
	}
	if src.calls != 1 {
		t.Fatalf("свежее значение: load вызван %d раз, ожидали 1", src.calls)
	}
	clk.t = clk.t.Add(time.Minute)
	src.set("v2", nil)
	if v, _ := c.Get("k"); v != "v2" || src.calls != 2 {
		t.Fatalf("после ttl ожидали перезагрузку: v=%q calls=%d", v, src.calls)
	}
}

func TestCacheStale(t *testing.T) {
	src := &chkSrc{val: "v1"}
	clk := &chkNow{time.Unix(0, 0)}
	c := NewCache(src.load, time.Minute, 10*time.Second, clk.now)
	c.Get("k")
	clk.t = clk.t.Add(90 * time.Second)
	src.set("", fmt.Errorf("timeout: %w", chkErrDown))
	v, err := c.Get("k")
	var se *StaleError
	if v != "v1" || !errors.As(err, &se) || se.Age != 90*time.Second || !errors.Is(err, chkErrDown) {
		t.Fatalf("источник упал: ожидали старое v1 и *StaleError{Age: 90s} с причиной, получили %q, %v", v, err)
	}
	c.Get("k")
	if src.calls != 3 {
		t.Fatalf("временную ошибку кэшировать нельзя: load вызван %d раз, ожидали 3", src.calls)
	}
	if v, err := c.Get("other"); v != "" || err == nil || errors.As(err, &se) || !errors.Is(err, chkErrDown) {
		t.Fatalf("без старого значения: ожидали \"\" и ошибку load как есть, получили %q, %v", v, err)
	}
}

func TestCacheNegative(t *testing.T) {
	src := &chkSrc{val: "v1"}
	clk := &chkNow{time.Unix(0, 0)}
	c := NewCache(src.load, time.Minute, 10*time.Second, clk.now)
	c.Get("k")
	clk.t = clk.t.Add(2 * time.Minute)
	src.set("", fmt.Errorf("user k: %w", ErrNotFound))
	for range 3 {
		if v, err := c.Get("k"); v != "" || !errors.Is(err, ErrNotFound) {
			t.Fatalf("удалённый ключ: ожидали \"\" и ErrNotFound (старое значение отдавать нельзя), получили %q, %v", v, err)
		}
	}
	if src.calls != 2 {
		t.Fatalf("«не найдено» кэшируется: load вызван %d раз, ожидали 2", src.calls)
	}
	clk.t = clk.t.Add(10 * time.Second)
	src.set("", chkErrDown)
	if v, err := c.Get("k"); v != "" || err != chkErrDown {
		t.Fatalf("после negTTL и падения источника: ожидали \"\" и ошибку load (старое значение забыто), получили %q, %v", v, err)
	}
}

func TestCacheLoadWithoutLock(t *testing.T) {
	started := make(chan struct{})
	load := func(k string) (string, error) {
		if k == "slow" {
			<-started // ждём, пока начнётся загрузка другого ключа
			return "s", nil
		}
		close(started)
		return "f", nil
	}
	clk := &chkNow{time.Unix(0, 0)}
	c := NewCache(load, time.Minute, time.Minute, clk.now)
	done := make(chan struct{})
	go func() { c.Get("slow"); close(done) }()
	time.Sleep(10 * time.Millisecond)
	go c.Get("fast")
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("загрузка одного ключа заблокировала другой: load вызывается под мьютексом")
	}
}
