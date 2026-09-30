package main

func chkAllowed(b *Bucket, n int) int {
	c := 0
	for range n {
		if b.Allow() {
			c++
		}
	}
	return c
}

func chkTick(t *testing.T, refill chan time.Time, what string) {
	t.Helper()
	select {
	case refill <- time.Now():
	case <-time.After(time.Second):
		t.Fatalf("%s: горутина пополнения не забрала тик — зависла на полном ведре?", what)
	}
}

func TestBucketStartsFull(t *testing.T) {
	b := NewBucket(3, make(chan time.Time))
	defer b.Stop()
	if got := chkAllowed(b, 10); got != 3 {
		t.Fatalf("из нового ведра на 3 выдано %d токенов, ожидали 3", got)
	}
}

func TestBucketRefillCapped(t *testing.T) {
	refill := make(chan time.Time)
	b := NewBucket(2, refill)
	defer b.Stop()
	for i := range 10 {
		chkTick(t, refill, fmt.Sprintf("тик %d в полное ведро", i+1))
	}
	if got := chkAllowed(b, 10); got != 2 {
		t.Fatalf("после 10 тиков в полное ведро на 2 выдано %d токенов, ожидали 2", got)
	}
	chkTick(t, refill, "тик в пустое ведро")
	chkTick(t, refill, "второй тик")
	time.Sleep(5 * time.Millisecond)
	if got := chkAllowed(b, 10); got != 2 {
		t.Fatalf("после двух тиков выдано %d, ожидали 2", got)
	}
}

func TestBucketWait(t *testing.T) {
	refill := make(chan time.Time)
	b := NewBucket(1, refill)
	defer b.Stop()
	b.Allow()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := b.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Wait на пустом ведре по таймауту = %v, ожидали DeadlineExceeded", err)
	}
	res := make(chan error, 1)
	go func() { res <- b.Wait(context.Background()) }()
	time.Sleep(5 * time.Millisecond)
	chkTick(t, refill, "тик для ждущего")
	select {
	case err := <-res:
		if err != nil {
			t.Fatalf("Wait после тика = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("тик пришёл, а Wait так и ждёт")
	}
}

func TestBucketConcurrentAllow(t *testing.T) {
	b := NewBucket(5, make(chan time.Time))
	defer b.Stop()
	var ok atomic.Int64
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Allow() {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()
	if ok.Load() != 5 {
		t.Fatalf("20 горутин получили %d токенов из ведра на 5", ok.Load())
	}
}

func TestBucketStopAndClosedRefill(t *testing.T) {
	before := runtime.NumGoroutine()
	b1 := NewBucket(1, make(chan time.Time))
	refill := make(chan time.Time)
	b2 := NewBucket(1, refill)
	stopped := make(chan struct{})
	go func() { b1.Stop(); b1.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop (дважды) не вернулся")
	}
	close(refill)
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после Stop и закрытия refill осталось %d горутин пополнения", n-before)
	}
	b2.Stop()
}
