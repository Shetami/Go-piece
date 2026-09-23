package main

func TestOrDoneForwards(t *testing.T) {
	in := make(chan int, 3)
	in <- 1
	in <- 2
	in <- 3
	close(in)
	var got []int
	for v := range OrDone(context.Background(), in) {
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("переслано %v, ожидали [1 2 3]", got)
	}
}

func TestOrDoneStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int) // никогда не закроется
	out := OrDone(ctx, in)
	cancel()
	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("после отмены пришло значение")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("после отмены выход не закрылся — range по нему висел бы вечно")
	}
}

func TestOrDoneNoLeakWhenReaderLeaves(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int, 1)
	in <- 1
	before := runtime.NumGoroutine()
	OrDone(ctx, in) // значение возьмут из in, но читать выход никто не будет
	time.Sleep(20 * time.Millisecond)
	cancel()
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после отмены осталась висящая горутина (%d > %d)", n, before)
	}
}
