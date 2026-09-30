package main

type chkClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkClock) Add(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

func chkWait[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(2 * time.Second):
		t.Fatalf("не дождались: %s", what)
	}
	var zero T
	return zero
}

func TestCacheNoStampede(t *testing.T) {
	clk := &chkClock{t: time.Unix(0, 0)}
	var calls atomic.Int32
	release := make(chan struct{})
	c := NewCache(time.Minute, clk.Now, func(ctx context.Context, k string) (int, error) {
		calls.Add(1)
		<-release
		return len(k), nil
	})
	res := make(chan int, 20)
	for range 20 {
		go func() { v, _ := c.Get(context.Background(), "user:42"); res <- v }()
	}
	time.Sleep(30 * time.Millisecond)
	close(release)
	for range 20 {
		if v := chkWait(t, res, "20 одновременных Get"); v != 7 {
			t.Fatalf("Get = %d, ожидали 7", v)
		}
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("20 одновременных промахов вызвали load %d раз, ожидали 1", n)
	}
	c.Get(context.Background(), "user:42")
	clk.Add(59 * time.Second)
	c.Get(context.Background(), "user:42")
	if n := calls.Load(); n != 1 {
		t.Fatalf("свежая запись перезагружена: load вызван %d раз", n)
	}
	clk.Add(time.Second)
	c.Get(context.Background(), "user:42")
	if n := calls.Load(); n != 2 {
		t.Fatalf("после ttl запись должна перезагрузиться: load вызван %d раз, ожидали 2", n)
	}
}

func TestCacheErrorNotCached(t *testing.T) {
	boom := errors.New("база недоступна")
	fail := true
	c := NewCache(time.Minute, time.Now, func(ctx context.Context, k int) (string, error) {
		if fail {
			return "", boom
		}
		return "ok", nil
	})
	if _, err := c.Get(context.Background(), 1); !errors.Is(err, boom) {
		t.Fatalf("Get = %v, ожидали ошибку load", err)
	}
	fail = false
	if v, err := c.Get(context.Background(), 1); err != nil || v != "ok" {
		t.Fatalf("после ошибки Get = %q, %v, ожидали повторную загрузку: ok, nil", v, err)
	}
}

func TestCacheKeysIndependent(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	c := NewCache(time.Minute, time.Now, func(ctx context.Context, k string) (string, error) {
		if k == "медленный" {
			<-block
		}
		return k, nil
	})
	go c.Get(context.Background(), "медленный")
	time.Sleep(20 * time.Millisecond)
	done := make(chan string, 1)
	go func() { v, _ := c.Get(context.Background(), "быстрый"); done <- v }()
	chkWait(t, done, "Get другого ключа, пока грузится медленный — load под общим мьютексом?")
}

func TestCacheCancel(t *testing.T) {
	release := make(chan struct{})
	var loadErr atomic.Value
	c := NewCache(time.Minute, time.Now, func(ctx context.Context, k string) (string, error) {
		<-release
		if err := ctx.Err(); err != nil {
			loadErr.Store(err)
			return "", err
		}
		return "v", nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() { _, err := c.Get(ctx, "k"); first <- err }()
	time.Sleep(20 * time.Millisecond)
	second := make(chan string, 1)
	go func() { v, _ := c.Get(context.Background(), "k"); second <- v }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	if err := chkWait(t, first, "Get с отменённым ctx должен вернуться, не дожидаясь загрузки"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get с отменённым ctx = %v, ожидали context.Canceled", err)
	}
	close(release)
	if v := chkWait(t, second, "второй Get после отмены первого"); v != "v" {
		t.Fatalf("второй Get = %q, ожидали v — отмена первого вызывающего сорвала загрузку (load увидел %v)", v, loadErr.Load())
	}
}
