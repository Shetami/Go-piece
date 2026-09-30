package main

type chkClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *chkClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *chkClock) Add(d time.Duration) { c.mu.Lock(); c.t = c.t.Add(d); c.mu.Unlock() }

func TestStoreExpiryWithoutCleanup(t *testing.T) {
	clk := &chkClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	s := NewStore(time.Minute, time.Hour, clk.Now)
	defer s.Close()
	s.Set("sid1", "alice")
	clk.Add(59 * time.Second)
	if u, ok := s.Get("sid1"); !ok || u != "alice" {
		t.Fatalf("через 59с при ttl 1м Get = %q, %v, ожидали alice, true", u, ok)
	}
	s.Set("sid1", "alice") // продление
	clk.Add(59 * time.Second)
	if _, ok := s.Get("sid1"); !ok {
		t.Fatal("Set должен продлевать сессию")
	}
	clk.Add(time.Second)
	if _, ok := s.Get("sid1"); ok {
		t.Fatal("ровно через ttl после Set сессия уже истекла, а Get её вернул (чистка ещё не прошла — проверяйте срок в Get)")
	}
	if _, ok := s.Get("нет-такой"); ok {
		t.Fatal("Get несуществующей сессии вернул true")
	}
}

func TestStoreCleanup(t *testing.T) {
	clk := &chkClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	s := NewStore(time.Minute, time.Millisecond, clk.Now)
	defer s.Close()
	for i := range 50 {
		s.Set(fmt.Sprint("sid", i), "u")
	}
	clk.Add(30 * time.Second)
	s.Set("fresh", "bob")
	clk.Add(40 * time.Second)
	deadline := time.Now().Add(3 * time.Second)
	for s.Len() != 1 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := s.Len(); n != 1 {
		t.Fatalf("фоновая чистка не удалила истёкшие сессии: Len() = %d, ожидали 1", n)
	}
	if u, ok := s.Get("fresh"); !ok || u != "bob" {
		t.Fatalf("чистка удалила живую сессию: %q, %v", u, ok)
	}
}

func TestStoreCloseNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 20 {
		s := NewStore(time.Minute, time.Millisecond, time.Now)
		s.Set("a", "b")
		s.Close()
		s.Close()
	}
	time.Sleep(20 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("после 20 NewStore+Close горутин стало %d, было %d — Close не останавливает чистку", after, before)
	}
}

func TestStoreConcurrent(t *testing.T) {
	s := NewStore(time.Minute, time.Millisecond, time.Now)
	defer s.Close()
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 300 {
				id := fmt.Sprint(g, "-", i%10)
				s.Set(id, "u")
				s.Get(id)
				s.Len()
			}
		}()
	}
	wg.Wait()
	if n := s.Len(); n != 80 {
		t.Fatalf("Len() = %d, ожидали 80", n)
	}
}
