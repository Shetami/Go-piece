package main

func chanOf(vals ...int) <-chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func TestBridgeOrder(t *testing.T) {
	streams := make(chan (<-chan int), 3)
	streams <- chanOf(1, 2)
	streams <- chanOf()
	streams <- chanOf(3, 4, 5)
	close(streams)

	var got []int
	for v := range Bridge(context.Background(), streams) {
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("получили %v, ожидали [1 2 3 4 5]", got)
	}
}

func TestBridgeEmpty(t *testing.T) {
	streams := make(chan (<-chan int))
	close(streams)
	select {
	case _, ok := <-Bridge(context.Background(), streams):
		if ok {
			t.Fatalf("из пустого потока пришло значение")
		}
	case <-time.After(time.Second):
		t.Fatalf("выход не закрылся")
	}
}

func TestBridgeCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	streams := make(chan (<-chan int)) // никогда не закрывается
	endless := make(chan int)          // тоже
	go func() { streams <- endless }()

	out := Bridge(ctx, streams)
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatalf("после отмены пришло значение")
		}
	case <-time.After(time.Second):
		t.Fatalf("после отмены выход не закрылся")
	}

	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после отмены осталось %d лишних горутин", n-before)
	}
}

func TestBridgeCancelWhileSending(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	streams := make(chan (<-chan int), 1)
	streams <- chanOf(1, 2, 3)
	out := Bridge(ctx, streams)
	<-out // берём одно и бросаем читать
	cancel()
	select {
	case <-time.After(time.Second):
		t.Fatalf("Bridge завис на отправке после отмены")
	case <-drainUntilClosed(out):
	}
}

func drainUntilClosed(ch <-chan int) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for range ch {
		}
		close(done)
	}()
	return done
}
