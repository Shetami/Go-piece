package main

func TestCounterConcurrent(t *testing.T) {
	c := NewCounter()
	keys := []string{"/", "/api", "/login"}
	var wg sync.WaitGroup
	for i := range 90 {
		wg.Go(func() {
			for range 100 {
				c.Inc(keys[i%3])
			}
		})
	}
	wg.Wait()
	for _, k := range keys {
		if got := c.Get(k); got != 3000 {
			t.Fatalf("Get(%q) = %d, ожидали 3000", k, got)
		}
	}
}

func TestCounterSnapshotIsCopy(t *testing.T) {
	c := NewCounter()
	c.Inc("a")
	snap := c.Snapshot()
	if snap["a"] != 1 {
		t.Fatalf("Snapshot = %v, ожидали a:1", snap)
	}
	snap["a"] = 100
	c.Inc("a")
	if c.Get("a") != 2 {
		t.Fatalf("изменение снимка повлияло на счётчик: Get(a) = %d", c.Get("a"))
	}
	if snap["a"] != 100 {
		t.Fatal("изменение счётчика повлияло на снимок")
	}
}

func TestCounterSnapshotDuringWrites(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 1000 {
			c.Inc("x")
		}
	})
	for range 100 {
		for range c.Snapshot() {
		}
	}
	wg.Wait()
}
