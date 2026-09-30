package main

func chkNaturals(stopped *atomic.Bool) iter.Seq[int] {
	return func(yield func(int) bool) {
		defer stopped.Store(true)
		for i := 0; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func TestValuesBreak(t *testing.T) {
	ch := make(chan int, 10)
	for i := 1; i <= 10; i++ {
		ch <- i
	}
	close(ch)
	var got []int
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("break в range по Values вызвал панику: %v", r)
			}
		}()
		for v := range Values(context.Background(), ch) {
			if v == 4 {
				break
			}
			got = append(got, v)
		}
	}()
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("до break получили %v, ожидали [1 2 3]", got)
	}
	if v := <-ch; v != 5 {
		t.Fatalf("после break следующее значение в канале %d, ожидали 5 — лишнее прочитано и потеряно", v)
	}
	rest := slices.Collect(Values(context.Background(), ch))
	if !reflect.DeepEqual(rest, []int{6, 7, 8, 9, 10}) {
		t.Fatalf("остаток %v, ожидали [6 7 8 9 10]", rest)
	}
}

func TestValuesCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan int) // никогда не закроется
	done := make(chan []int, 1)
	go func() { done <- slices.Collect(Values(ctx, ch)) }()
	select {
	case ch <- 1:
	case <-time.After(time.Second):
		t.Fatal("range по Values не читает канал")
	}
	cancel()
	select {
	case got := <-done:
		if !reflect.DeepEqual(got, []int{1}) {
			t.Fatalf("получили %v, ожидали [1]", got)
		}
	case <-time.After(time.Second):
		t.Fatal("range по Values не завершился после отмены ctx")
	}
}

func TestStreamRoundTrip(t *testing.T) {
	out := Stream(context.Background(), slices.Values([]string{"a", "", "c"}))
	if out == nil {
		t.Fatal("Stream вернул nil-канал")
	}
	var got []string
	timeout := time.After(time.Second)
	for {
		select {
		case v, ok := <-out:
			if !ok {
				if !reflect.DeepEqual(got, []string{"a", "", "c"}) {
					t.Fatalf("получили %q, ожидали [a  c]", got)
				}
				return
			}
			got = append(got, v)
		case <-timeout:
			t.Fatalf("канал не закрылся после конца seq, получено %q", got)
		}
	}
}

func TestStreamCancelStopsSeq(t *testing.T) {
	before := runtime.NumGoroutine()
	var stopped atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())
	out := Stream(ctx, chkNaturals(&stopped))
	if out == nil {
		t.Fatal("Stream вернул nil-канал")
	}
	for want := range 3 {
		if v := <-out; v != want {
			t.Fatalf("получили %d, ожидали %d", v, want)
		}
	}
	cancel() // дальше выход никто не читает
	deadline := time.Now().Add(time.Second)
	for (!stopped.Load() || runtime.NumGoroutine() > before) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !stopped.Load() {
		t.Fatal("после отмены бесконечный seq так и не остановился")
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после отмены осталось %d лишних горутин", n-before)
	}
	for range out {
	}
}
