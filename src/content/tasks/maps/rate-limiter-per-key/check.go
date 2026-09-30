package main

type chkLimClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkLimClock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkLimClock) add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func chkAllowN(l *Limiter, key string, n int) int {
	ok := 0
	for range n {
		if l.Allow(key) {
			ok++
		}
	}
	return ok
}

func TestLimiterBurstAndKeys(t *testing.T) {
	clk := &chkLimClock{t: time.Unix(1_700_000_000, 0)}
	l := NewLimiter(1, 3, clk.now)
	if n := chkAllowN(l, "u1", 5); n != 3 {
		t.Fatalf("новый ключ с burst 3: пропущено %d из 5, ожидали 3", n)
	}
	if n := chkAllowN(l, "u2", 5); n != 3 {
		t.Fatalf("второй ключ пропущено %d из 5, ожидали 3 — вёдра ключей независимы", n)
	}
	clk.add(time.Hour)
	if n := chkAllowN(l, "u1", 10); n != 3 {
		t.Fatalf("после часа простоя пропущено %d, ожидали 3 — ведро не больше burst", n)
	}
}

func TestLimiterFractionalRefill(t *testing.T) {
	clk := &chkLimClock{t: time.Unix(1_700_000_000, 0)}
	l := NewLimiter(2, 2, clk.now) // токен каждые 500мс
	allowed := 0
	for range 20 { // 20 запросов раз в 300мс: 2 из ведра + ≈11 долитых за 5.7с
		if l.Allow("k") {
			allowed++
		}
		clk.add(300 * time.Millisecond)
	}
	if allowed < 12 || allowed > 14 {
		t.Fatalf("за 6с при rate 2/с пропущено %d, ожидали около 13 — дробные доли токена теряются", allowed)
	}
}

func TestLimiterCleanup(t *testing.T) {
	clk := &chkLimClock{t: time.Unix(1_700_000_000, 0)}
	l := NewLimiter(1, 2, clk.now)
	l.Allow("idle")
	clk.add(10 * time.Second)
	chkAllowN(l, "busy", 2)
	if n := l.Cleanup(); n != 1 || l.Len() != 1 {
		t.Fatalf("Cleanup = %d, Len = %d; ожидали 1 и 1: idle уже полон, busy — нет", n, l.Len())
	}
	if l.Allow("busy") {
		t.Fatal("Cleanup сбросил пустое ведро busy — его нельзя удалять")
	}
}

func TestLimiterConcurrent(t *testing.T) {
	clk := &chkLimClock{t: time.Unix(1_700_000_000, 0)}
	l := NewLimiter(0.001, 10, clk.now)
	var ok atomic.Int32
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 4 {
				if l.Allow("shared") {
					ok.Add(1)
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()
	if n := ok.Load(); n != 10 {
		t.Fatalf("200 параллельных запросов при burst 10: пропущено %d, ожидали ровно 10", n)
	}
}
