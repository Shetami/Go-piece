package main

func TestCounterBasic(t *testing.T) {
	c := NewCounter()
	c.Add("GET /", 2)
	c.Add("GET /", 3)
	c.Add("POST /login", 1)
	if g := c.Get("GET /"); g != 5 {
		t.Fatalf("Get(GET /) = %d, ожидали 5", g)
	}
	d := c.Drain()
	if len(d) != 2 || d["GET /"] != 5 || d["POST /login"] != 1 {
		t.Fatalf("Drain() = %v, ожидали map[GET /:5 POST /login:1]", d)
	}
	if g := c.Get("GET /"); g != 0 {
		t.Fatalf("после Drain Get = %d, ожидали 0", g)
	}
	d["GET /"] = 100
	if d2 := c.Drain(); d2 == nil || len(d2) != 0 {
		t.Fatalf("второй Drain без новых Add = %v, ожидали пустую не-nil мапу", d2)
	}
	if g := c.Get("GET /"); g != 0 {
		t.Fatalf("изменение результата Drain повлияло на счётчик: Get = %d", g)
	}
}

func TestCounterDrainNoLoss(t *testing.T) {
	const writers, perWriter = 8, 3000
	c := NewCounter()
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	totals := map[string]int64{}
	var wg sync.WaitGroup
	for w := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range perWriter {
				c.Add(keys[(i+w)%len(keys)], 1)
				if i%64 == 0 {
					runtime.Gosched()
				}
			}
		}()
	}
	stop := make(chan struct{})
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		for {
			for k, v := range c.Drain() {
				totals[k] += v
			}
			select {
			case <-stop:
				return
			default:
				runtime.Gosched()
			}
		}
	}()
	wg.Wait()
	close(stop)
	<-drained
	for k, v := range c.Drain() {
		totals[k] += v
	}
	var sum int64
	for _, v := range totals {
		sum += v
	}
	if sum != writers*perWriter {
		t.Fatalf("сумма всех Drain = %d, ожидали %d — прибавления, пришедшие во время Drain, потеряны или посчитаны дважды", sum, writers*perWriter)
	}
	for _, k := range keys {
		if totals[k] != writers*perWriter/int64(len(keys)) {
			t.Fatalf("ключ %q: %d, ожидали %d", k, totals[k], writers*perWriter/len(keys))
		}
	}
}
