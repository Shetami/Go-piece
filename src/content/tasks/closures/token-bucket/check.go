package main

type chkClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *chkClock) Add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func chkNewClock() *chkClock { return &chkClock{t: time.Unix(1_700_000_000, 0)} }

func chkCount(allow func() bool, n int) int {
	ok := 0
	for range n {
		if allow() {
			ok++
		}
	}
	return ok
}

func TestLimiterBurstAndRefill(t *testing.T) {
	c := chkNewClock()
	allow := NewLimiter(1, 3, c.Now)
	if got := chkCount(allow, 5); got != 3 {
		t.Fatalf("полный бакет на 3: пропущено %d из 5, ожидали 3", got)
	}
	c.Add(time.Second)
	if got := chkCount(allow, 5); got != 1 {
		t.Fatalf("через секунду при rate=1: пропущено %d, ожидали 1", got)
	}
	c.Add(time.Hour)
	if got := chkCount(allow, 10); got != 3 {
		t.Fatalf("после часа простоя: пропущено %d, ожидали 3 — бакет не копит больше burst", got)
	}
}

func TestLimiterFractions(t *testing.T) {
	c := chkNewClock()
	allow := NewLimiter(2, 1, c.Now) // полтокена каждые 250 мс
	allow()
	var got []bool
	for range 6 {
		c.Add(250 * time.Millisecond)
		got = append(got, allow())
	}
	want := []bool{false, true, false, true, false, true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("вызовы каждые 250 мс при rate=2: %v, ожидали %v — доли токена должны копиться", got, want)
	}
}

func TestLimiterClockBackwards(t *testing.T) {
	c := chkNewClock()
	allow := NewLimiter(1, 1, c.Now)
	allow()
	c.Add(-10 * time.Second)
	if allow() {
		t.Fatal("часы ушли назад — токенов прибавиться не должно")
	}
	c.Add(10 * time.Second)
	if allow() {
		t.Fatal("часы вернулись в исходную точку — время на самом деле не шло, токена нет")
	}
	c.Add(time.Second)
	if !allow() {
		t.Fatal("прошла секунда вперёд — токен должен появиться")
	}
}

func TestLimiterConcurrent(t *testing.T) {
	c := chkNewClock()
	allow := NewLimiter(1, 10, c.Now)
	var ok atomic.Int32
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if allow() {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if n := ok.Load(); n != 10 {
		t.Fatalf("100 горутин, burst 10: пропущено %d, ожидали ровно 10", n)
	}
}
