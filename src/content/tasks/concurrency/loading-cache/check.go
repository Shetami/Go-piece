package main

// chkClock — ручные часы для тестов.
type chkClock struct{ ns atomic.Int64 }

func (c *chkClock) now() time.Time          { return time.Unix(0, c.ns.Load()) }
func (c *chkClock) advance(d time.Duration) { c.ns.Add(int64(d)) }

func TestCacheHitAndTTL(t *testing.T) {
	var clk chkClock
	calls := 0
	c := NewCache(time.Minute, clk.now, func(k string) (string, error) {
		calls++
		return fmt.Sprintf("%s#%d", k, calls), nil
	})
	ctx := context.Background()
	v1, _ := c.Get(ctx, "a")
	clk.advance(59 * time.Second)
	v2, _ := c.Get(ctx, "a")
	if v1 != "a#1" || v2 != "a#1" {
		t.Fatalf("Get = %q, затем %q; ожидали \"a#1\" дважды — свежее значение берётся из кэша", v1, v2)
	}
	clk.advance(time.Second)
	if v3, _ := c.Get(ctx, "a"); v3 != "a#2" {
		t.Fatalf("после ttl Get = %q, ожидали перезагрузку \"a#2\"", v3)
	}
}

func TestCacheErrorsNotCached(t *testing.T) {
	var clk chkClock
	fail := true
	c := NewCache(time.Hour, clk.now, func(k int) (int, error) {
		if fail {
			return 0, errors.New("timeout")
		}
		return k * 10, nil
	})
	if _, err := c.Get(context.Background(), 1); err == nil {
		t.Fatal("load вернул ошибку, а Get — nil")
	}
	fail = false
	if v, err := c.Get(context.Background(), 1); err != nil || v != 10 {
		t.Fatalf("после ошибки Get = %d, %v; ожидали 10, nil — ошибка не должна кэшироваться", v, err)
	}
}

func TestCacheConcurrentLoadOnce(t *testing.T) {
	var clk chkClock
	var calls atomic.Int32
	c := NewCache(time.Hour, clk.now, func(k string) (int, error) {
		calls.Add(1)
		time.Sleep(30 * time.Millisecond)
		return 42, nil
	})
	var wg sync.WaitGroup
	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := c.Get(context.Background(), "hot"); v != 42 || err != nil {
				t.Errorf("Get = %d, %v; ожидали 42, nil", v, err)
			}
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("30 одновременных Get вызвали load %d раз, ожидали 1", n)
	}
}

func TestCacheKeysIndependent(t *testing.T) {
	var clk chkClock
	release := make(chan struct{})
	defer close(release)
	c := NewCache(time.Hour, clk.now, func(k string) (string, error) {
		if k == "slow" {
			<-release
		}
		return k, nil
	})
	go c.Get(context.Background(), "slow")
	time.Sleep(20 * time.Millisecond)
	done := make(chan string, 1)
	go func() { v, _ := c.Get(context.Background(), "fast"); done <- v }()
	select {
	case v := <-done:
		if v != "fast" {
			t.Fatalf("Get(\"fast\") = %q", v)
		}
	case <-time.After(time.Second):
		t.Fatal("Get(\"fast\") ждёт медленную загрузку другого ключа")
	}
}

func TestCacheWaiterCancel(t *testing.T) {
	var clk chkClock
	var calls atomic.Int32
	release := make(chan struct{})
	c := NewCache(time.Hour, clk.now, func(k string) (string, error) {
		calls.Add(1)
		<-release
		return "loaded", nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	res := make(chan error, 1)
	go func() { _, err := c.Get(ctx, "k"); res <- err }()
	select {
	case err := <-res:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Get с истёкшим ctx вернул %v, ожидали DeadlineExceeded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Get не вернулся по истечении ctx — ждёт загрузку")
	}
	close(release)
	if v, err := c.Get(context.Background(), "k"); v != "loaded" || err != nil {
		t.Fatalf("Get = %q, %v; ожидали \"loaded\", nil", v, err)
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("load вызван %d раз, ожидали 1 — брошенная загрузка должна попасть в кэш", n)
	}
}
