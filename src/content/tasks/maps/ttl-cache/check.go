package main

type chkClock struct{ t time.Time }

func (c *chkClock) now() time.Time      { return c.t }
func (c *chkClock) add(d time.Duration) { c.t = c.t.Add(d) }
func newChkClock() *chkClock            { return &chkClock{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)} }

func TestTTLExpiryBoundary(t *testing.T) {
	clk := newChkClock()
	c := NewTTLCache[string, int](clk.now)
	c.Set("a", 1, 10*time.Second)
	clk.add(9 * time.Second)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("через 9с при ttl 10с: Get = %d, %v; ожидали 1, true", v, ok)
	}
	clk.add(time.Second) // ровно 10с — уже истекла
	if _, ok := c.Get("a"); ok {
		t.Fatal("ровно в момент истечения запись должна быть мёртвой")
	}
}

func TestTTLForeverAndLen(t *testing.T) {
	clk := newChkClock()
	c := NewTTLCache[string, int](clk.now)
	c.Set("forever", 1, 0)
	c.Set("neg", 2, -time.Second)
	c.Set("short", 3, time.Second)
	c.Set("long", 4, time.Hour)
	clk.add(time.Minute)
	if n := c.Len(); n != 3 {
		t.Fatalf("Len через минуту = %d, ожидали 3 (истёкшую short не считаем, даже без Get)", n)
	}
	clk.add(1000 * time.Hour)
	if _, ok := c.Get("forever"); !ok {
		t.Fatal("запись с ttl 0 истекла")
	}
	if _, ok := c.Get("neg"); !ok {
		t.Fatal("запись с отрицательным ttl должна быть бессрочной")
	}
}

func TestTTLResetExtends(t *testing.T) {
	clk := newChkClock()
	c := NewTTLCache[string, string](clk.now)
	c.Set("a", "old", time.Second)
	c.Set("a", "new", 10*time.Second)
	clk.add(5 * time.Second)
	if n := c.Purge(); n != 0 {
		t.Fatalf("Purge = %d, ожидали 0: первый срок ключа a устарел после повторного Set", n)
	}
	if v, ok := c.Get("a"); !ok || v != "new" {
		t.Fatalf("Get(a) = %q, %v; ожидали new, true", v, ok)
	}
	c.Set("a", "short", time.Second) // укоротили
	clk.add(2 * time.Second)
	if _, ok := c.Get("a"); ok {
		t.Fatal("повторный Set с коротким ttl должен укоротить срок")
	}
}

func TestTTLPurgeCount(t *testing.T) {
	clk := newChkClock()
	c := NewTTLCache[int, int](clk.now)
	for i := range 100 {
		c.Set(i, i, time.Duration(i+1)*time.Second)
	}
	clk.add(30 * time.Second)
	c.Get(5) // уже удалена через Get — Purge её не считает
	if n := c.Purge(); n != 29 {
		t.Fatalf("Purge = %d, ожидали 29", n)
	}
	if n := c.Purge(); n != 0 {
		t.Fatalf("повторный Purge = %d, ожидали 0", n)
	}
	if n := c.Len(); n != 70 {
		t.Fatalf("Len = %d, ожидали 70", n)
	}
}

func TestTTLConcurrent(t *testing.T) {
	clk := newChkClock()
	c := NewTTLCache[int, int](clk.now)
	var wg sync.WaitGroup
	for g := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 200 {
				c.Set(i%20, g, time.Duration(i%3)*time.Second)
				c.Get(i % 20)
				c.Len()
			}
		}()
	}
	wg.Wait()
}
