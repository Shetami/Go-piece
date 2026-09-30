package main

func chkRunPriority(t *testing.T, ctx context.Context, workers int, high, low <-chan func()) error {
	t.Helper()
	res := make(chan error, 1)
	go func() { res <- RunPriority(ctx, workers, high, low) }()
	select {
	case err := <-res:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("RunPriority не вернулся")
		return nil
	}
}

func TestPriorityHighFirst(t *testing.T) {
	high := make(chan func(), 10)
	low := make(chan func(), 10)
	var mu sync.Mutex
	var order []string
	add := func(s string) func() { return func() { mu.Lock(); order = append(order, s); mu.Unlock() } }
	for i := range 10 {
		low <- add(fmt.Sprint("low", i))
	}
	for i := range 10 {
		high <- add(fmt.Sprint("high", i))
	}
	close(high)
	close(low)
	if err := chkRunPriority(t, context.Background(), 1, high, low); err != nil {
		t.Fatalf("RunPriority = %v, ожидали nil", err)
	}
	if len(order) != 20 {
		t.Fatalf("выполнено %d задач из 20", len(order))
	}
	for i := range 10 {
		if !strings.HasPrefix(order[i], "high") {
			t.Fatalf("порядок %v: задача из low выполнена, пока в high ещё были задачи", order)
		}
	}
}

func TestPriorityOneQueueClosedEarly(t *testing.T) {
	high := make(chan func())
	low := make(chan func())
	close(high) // срочных задач не будет вовсе
	var done atomic.Int32
	go func() {
		for range 50 {
			low <- func() { done.Add(1) }
		}
		close(low)
	}()
	if err := chkRunPriority(t, context.Background(), 3, high, low); err != nil {
		t.Fatalf("RunPriority = %v", err)
	}
	if n := done.Load(); n != 50 {
		t.Fatalf("выполнено %d задач из low, ожидали 50 — закрытая high не должна останавливать работу", n)
	}
}

func TestPriorityWaitsForBothQueues(t *testing.T) {
	high := make(chan func())
	low := make(chan func())
	res := make(chan error, 1)
	go func() { res <- RunPriority(context.Background(), 2, high, low) }()
	close(low)
	select {
	case <-res:
		t.Fatal("RunPriority вернулся, хотя high ещё открыта")
	case <-time.After(50 * time.Millisecond):
	}
	var ran atomic.Bool
	high <- func() { ran.Store(true) }
	close(high)
	if err := <-res; err != nil || !ran.Load() {
		t.Fatalf("RunPriority = %v, задача выполнена = %v", err, ran.Load())
	}
}

func TestPriorityCancel(t *testing.T) {
	base := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	high := make(chan func(), 100)
	low := make(chan func(), 100)
	var started, finished atomic.Int32
	for range 100 {
		low <- func() {
			started.Add(1)
			time.Sleep(20 * time.Millisecond)
			finished.Add(1)
		}
	}
	go func() { time.Sleep(30 * time.Millisecond); cancel() }()
	err := chkRunPriority(t, ctx, 2, high, low)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("после отмены RunPriority = %v, ожидали context.Canceled", err)
	}
	if s, f := started.Load(), finished.Load(); s != f {
		t.Fatalf("RunPriority вернулся, пока %d задач ещё выполнялись", s-f)
	}
	if s := started.Load(); s > 8 {
		t.Fatalf("после отмены начато ещё много задач: всего %d", s)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if d := runtime.NumGoroutine() - base; d > 0 {
		t.Fatalf("осталось %d лишних горутин", d)
	}
}
