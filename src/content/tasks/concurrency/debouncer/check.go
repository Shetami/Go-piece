package main

type chkCalls struct {
	mu   sync.Mutex
	vals []int
}

func (c *chkCalls) fn(v int) { c.mu.Lock(); c.vals = append(c.vals, v); c.mu.Unlock() }
func (c *chkCalls) get() []int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.vals)
}

func TestDebounceBurst(t *testing.T) {
	var c chkCalls
	d := NewDebouncer(100*time.Millisecond, c.fn)
	defer d.Stop()
	for i := 1; i <= 5; i++ {
		d.Trigger(i)
		time.Sleep(20 * time.Millisecond)
	}
	if got := c.get(); len(got) != 0 {
		t.Fatalf("fn вызвана %v посреди серии событий — отсчёт должен идти от последнего Trigger", got)
	}
	time.Sleep(400 * time.Millisecond)
	if got := c.get(); !reflect.DeepEqual(got, []int{5}) {
		t.Fatalf("после серии fn вызвана с %v, ожидали один вызов [5]", got)
	}
	d.Trigger(6)
	time.Sleep(400 * time.Millisecond)
	if got := c.get(); !reflect.DeepEqual(got, []int{5, 6}) {
		t.Fatalf("после второй серии вызовы %v, ожидали [5 6]", got)
	}
}

func TestDebounceFlush(t *testing.T) {
	var c chkCalls
	d := NewDebouncer(100*time.Millisecond, c.fn)
	defer d.Stop()
	d.Flush() // нечего отправлять
	d.Trigger(1)
	d.Trigger(2)
	d.Flush()
	if got := c.get(); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("сразу после Flush вызовы %v, ожидали [2]", got)
	}
	time.Sleep(300 * time.Millisecond)
	if got := c.get(); !reflect.DeepEqual(got, []int{2}) {
		t.Fatalf("после Flush отложенный вызов всё равно сработал: %v", got)
	}
}

func TestDebounceStop(t *testing.T) {
	var c chkCalls
	d := NewDebouncer(50*time.Millisecond, c.fn)
	d.Trigger(1)
	d.Stop()
	d.Trigger(2)
	d.Flush()
	time.Sleep(200 * time.Millisecond)
	if got := c.get(); len(got) != 0 {
		t.Fatalf("после Stop fn вызвана с %v", got)
	}
}

func TestDebounceNoOverlapAndOrder(t *testing.T) {
	var active atomic.Int32
	var mu sync.Mutex
	var seen []int
	d := NewDebouncer(time.Millisecond, func(v int) {
		if active.Add(1) > 1 {
			t.Errorf("вызовы fn перекрылись")
		}
		time.Sleep(3 * time.Millisecond)
		mu.Lock()
		seen = append(seen, v)
		mu.Unlock()
		active.Add(-1)
	})
	var next atomic.Int32
	var tmu sync.Mutex // значения растут в порядке Trigger
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 30 {
				tmu.Lock()
				d.Trigger(int(next.Add(1)))
				tmu.Unlock()
				time.Sleep(time.Millisecond)
			}
		}()
	}
	wg.Wait()
	time.Sleep(50 * time.Millisecond)
	d.Stop()
	mu.Lock()
	defer mu.Unlock()
	if len(seen) == 0 || !slices.IsSorted(seen) {
		t.Fatalf("fn получила значения не по порядку: %v", seen)
	}
	if seen[len(seen)-1] != int(next.Load()) {
		t.Fatalf("последний вызов fn со значением %d, ожидали последнее %d", seen[len(seen)-1], next.Load())
	}
}
