package main

func chkMs(ms int) time.Time {
	return time.Unix(1_700_000_000, 0).Add(time.Duration(ms) * time.Millisecond)
}

func TestThrottleLeadingTrailing(t *testing.T) {
	var got []int
	call, tick := NewThrottle(100*time.Millisecond, func(v int) { got = append(got, v) })
	call(1, chkMs(0))
	call(2, chkMs(10))
	call(3, chkMs(20))
	tick(chkMs(50))
	tick(chkMs(100))
	call(4, chkMs(150))
	tick(chkMs(199))
	tick(chkMs(200))
	tick(chkMs(400))
	if want := []int{1, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Fatalf("f получила %v, ожидали %v: первый сразу, отложенный — последний, и отсчёт интервала от отложенного вызова", got, want)
	}
}

func TestThrottleStalePending(t *testing.T) {
	var got []int
	call, tick := NewThrottle(100*time.Millisecond, func(v int) { got = append(got, v) })
	call(1, chkMs(0))
	call(2, chkMs(50))
	call(3, chkMs(120))
	tick(chkMs(300))
	if want := []int{1, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("f получила %v, ожидали %v: отложенное 2 устарело, когда пришло 3 после интервала", got, want)
	}
}

func TestThrottleZeroValue(t *testing.T) {
	calls := 0
	call, tick := NewThrottle(100*time.Millisecond, func(s string) { calls++ })
	call("", chkMs(0))
	call("", chkMs(10))
	tick(chkMs(100))
	if calls != 2 {
		t.Fatalf("отложенная пустая строка — тоже значение: вызовов f %d, ожидали 2", calls)
	}
}

func TestThrottleReentrant(t *testing.T) {
	var got []int
	var now time.Time
	var call func(int, time.Time)
	var tick func(time.Time)
	call, tick = NewThrottle(100*time.Millisecond, func(v int) {
		got = append(got, v)
		if v < 3 {
			call(v+1, now) // f сама планирует следующую порцию
		}
	})
	done := make(chan struct{})
	go func() {
		defer close(done)
		now = chkMs(0)
		call(1, now)
		now = chkMs(100)
		tick(now)
		now = chkMs(200)
		tick(now)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("дедлок: f вызывает call, а call ждёт замок, который держит сам себя — f нельзя вызывать под блокировкой")
	}
	if want := []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("f получила %v, ожидали %v", got, want)
	}
}

func TestThrottleConcurrent(t *testing.T) {
	var calls atomic.Int32
	call, tick := NewThrottle(time.Second, func(int) { calls.Add(1) })
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			call(i, chkMs(0))
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("50 одновременных call: f вызвана %d раз, ожидали 1", n)
	}
	tick(chkMs(1000))
	tick(chkMs(1000))
	if n := calls.Load(); n != 2 {
		t.Fatalf("после интервала отложенное должно уйти ровно один раз: всего вызовов %d, ожидали 2", n)
	}
}
