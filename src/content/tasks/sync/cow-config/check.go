package main

func TestHolderSnapshotsImmutable(t *testing.T) {
	src := map[string]int{"free": 10}
	h := NewHolder(Config{Version: 1, Limits: src})
	src["free"] = 999
	old := h.Load()
	if old.Limits["free"] != 10 {
		t.Fatalf("вызывающий поменял свою мапу после NewHolder — и конфиг изменился: %v", old.Limits)
	}
	h.Update(func(c *Config) { c.Limits["free"] = 20; c.Limits["pro"] = 100 })
	if old.Version != 1 || old.Limits["free"] != 10 || len(old.Limits) != 1 {
		t.Fatalf("старый снимок изменился после Update: %+v — копия конфига делит мапу с оригиналом", *old)
	}
	cur := h.Load()
	if cur.Version != 2 || cur.Limits["free"] != 20 || cur.Limits["pro"] != 100 {
		t.Fatalf("после Update Load() = %+v, ожидали Version 2 и лимиты free:20 pro:100", *cur)
	}
}

func TestHolderConcurrentUpdates(t *testing.T) {
	h := NewHolder(Config{Limits: map[string]int{}})
	stop := make(chan struct{})
	var readers sync.WaitGroup
	for range 4 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				c := h.Load()
				if len(c.Limits) != c.Version {
					t.Errorf("снимок версии %d содержит %d ключей — читатель увидел недописанный конфиг", c.Version, len(c.Limits))
					return
				}
			}
		}()
	}
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Update(func(c *Config) { c.Limits[fmt.Sprint("client-", i)] = i })
		}()
	}
	wg.Wait()
	close(stop)
	readers.Wait()
	c := h.Load()
	if c.Version != 50 || len(c.Limits) != 50 {
		t.Fatalf("после 50 конкурентных Update: Version %d, ключей %d, ожидали 50 и 50 — обновления потеряны", c.Version, len(c.Limits))
	}
}

func TestHolderLoadNeverWaits(t *testing.T) {
	h := NewHolder(Config{Version: 7})
	inside := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	go h.Update(func(c *Config) {
		once.Do(func() { close(inside) })
		<-release
	})
	<-inside
	got := make(chan int, 1)
	go func() { got <- h.Load().Version }()
	select {
	case v := <-got:
		if v != 7 {
			t.Fatalf("Load во время Update вернул версию %d, ожидали старую 7", v)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Load ждёт, пока выполняется fn в Update — читатели не должны блокироваться")
	}
	close(release)
}
