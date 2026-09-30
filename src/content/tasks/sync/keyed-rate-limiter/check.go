package main

type chkClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClock) Now() time.Time      { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkClock) Add(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

func chkAllowed(l *Limiter, key string, n int) int {
	ok := 0
	for range n {
		if l.Allow(key) {
			ok++
		}
	}
	return ok
}

func TestLimiterBurstAndKeys(t *testing.T) {
	clk := &chkClock{t: time.Unix(1000, 0)}
	l := NewLimiter(1, 3, clk.Now)
	if n := chkAllowed(l, "alice", 5); n != 3 {
		t.Fatalf("новый клиент: пропущено %d из 5, ожидали burst = 3", n)
	}
	if n := chkAllowed(l, "bob", 5); n != 3 {
		t.Fatalf("bob пропущено %d из 5 — вёдра клиентов не независимы", n)
	}
	clk.Add(time.Hour)
	if n := chkAllowed(l, "alice", 10); n != 3 {
		t.Fatalf("после часа простоя пропущено %d, ожидали не больше burst = 3", n)
	}
}

func TestLimiterFractionalRefill(t *testing.T) {
	clk := &chkClock{t: time.Unix(1000, 0)}
	l := NewLimiter(1, 1, clk.Now) // 1 токен в секунду
	l.Allow("k")
	got := 0
	for range 10 {
		clk.Add(500 * time.Millisecond)
		if l.Allow("k") {
			got++
		}
	}
	// За 5 секунд накапливается 5 токенов.
	if got != 5 {
		t.Fatalf("запросы каждые 500мс при 1 токене/с: пропущено %d из 10 за 5с, ожидали 5 — дробные токены теряются?", got)
	}
}

func TestLimiterClockBackwards(t *testing.T) {
	clk := &chkClock{t: time.Unix(1000, 0)}
	l := NewLimiter(1, 2, clk.Now)
	chkAllowed(l, "k", 2)
	clk.Add(-time.Minute)
	if l.Allow("k") {
		t.Fatal("часы пошли назад — токенов не должно прибавиться")
	}
	clk.Add(time.Minute + time.Second)
	if n := chkAllowed(l, "k", 3); n != 1 {
		t.Fatalf("через секунду после исходного момента пропущено %d, ожидали 1", n)
	}
}

func TestLimiterCleanup(t *testing.T) {
	clk := &chkClock{t: time.Unix(1000, 0)}
	l := NewLimiter(1, 2, clk.Now)
	l.Allow("idle")
	chkAllowed(l, "busy", 2)
	clk.Add(1500 * time.Millisecond)
	if n := l.Cleanup(); n != 1 || l.Len() != 1 {
		t.Fatalf("Cleanup() = %d, Len() = %d; ожидали удалить только пополнившееся ведро idle", n, l.Len())
	}
	if n := chkAllowed(l, "busy", 3); n != 1 {
		t.Fatalf("Cleanup обнулил или удалил неполное ведро busy: пропущено %d, ожидали 1", n)
	}
}

func TestLimiterConcurrent(t *testing.T) {
	clk := &chkClock{t: time.Unix(1000, 0)}
	l := NewLimiter(10, 50, clk.Now)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				if l.Allow("api-key") {
					allowed.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	if n := allowed.Load(); n != 50 {
		t.Fatalf("400 одновременных запросов при замороженных часах: пропущено %d, ожидали ровно burst = 50", n)
	}
}
