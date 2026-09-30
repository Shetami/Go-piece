package main

func chkLeakWorker(ch chan struct{}, done *sync.WaitGroup) {
	defer done.Done()
	<-ch
}

func TestLeakNone(t *testing.T) {
	check := Snapshot()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done() }()
	}
	wg.Wait()
	if err := check(time.Second); err != nil {
		t.Fatalf("утечки нет, а check вернула ошибку: %.200s", err)
	}
}

func TestLeakLateExit(t *testing.T) {
	check := Snapshot()
	for i := 0; i < 3; i++ {
		go func() { time.Sleep(30 * time.Millisecond) }()
	}
	if err := check(2 * time.Second); err != nil {
		t.Fatalf("горутины завершаются через 30 мс, timeout 2 с — это не утечка, а check вернула: %.200s", err)
	}
}

func TestLeakDetected(t *testing.T) {
	const leaked = 800
	check := Snapshot()
	ch := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < leaked; i++ {
		wg.Add(1)
		go chkLeakWorker(ch, &wg)
	}
	defer func() { close(ch); wg.Wait() }()

	res := make(chan error, 1)
	go func() { res <- check(50 * time.Millisecond) }()
	var err error
	select {
	case err = <-res:
	case <-time.After(10 * time.Second):
		t.Fatalf("check(50ms) не вернулась за 10 с")
	}
	if err == nil {
		t.Fatalf("%d горутин висят на канале, а check вернула nil", leaked)
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "goroutine leak") {
		t.Fatalf("текст ошибки должен начинаться с \"goroutine leak\": %.100q", msg)
	}
	if n := strings.Count(msg, "chkLeakWorker("); n < leaked {
		t.Fatalf("в дампе %d стеков chkLeakWorker из %d — дамп обрезан (runtime.Stack пишет не больше размера буфера)", n, leaked)
	}
}

func TestLeakCheckItselfClean(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 20; i++ {
		check := Snapshot()
		ch := make(chan struct{})
		go func() { <-ch }()
		_ = check(time.Millisecond)
		close(ch)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > before; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после 20 вызовов check горутин стало %d, было %d — check оставляет свои горутины", n, before)
	}
}
