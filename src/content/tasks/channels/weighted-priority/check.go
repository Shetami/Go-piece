package main

func chkFilled(vals ...string) chan string {
	ch := make(chan string, len(vals))
	for _, v := range vals {
		ch <- v
	}
	return ch
}

func chkDrain(t *testing.T, out <-chan string) []string {
	t.Helper()
	if out == nil {
		t.Fatal("Prioritize вернул nil-канал")
	}
	var got []string
	for {
		select {
		case v, ok := <-out:
			if !ok {
				return got
			}
			got = append(got, v)
		case <-time.After(2 * time.Second):
			t.Fatalf("выход не закрылся, получено %v", got)
		}
	}
}

func TestPrioritizeWeighted(t *testing.T) {
	for range 20 {
		high := chkFilled("h0", "h1", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9")
		low := chkFilled("l0", "l1", "l2")
		close(high)
		close(low)
		got := chkDrain(t, Prioritize(context.Background(), high, low, 3))
		want := []string{"h0", "h1", "h2", "l0", "h3", "h4", "h5", "l1", "h6", "h7", "h8", "l2", "h9"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("получили %v,\nожидали %v (три из high, одно из low, и так далее)", got, want)
		}
	}
}

func TestPrioritizeLowWhenHighIdle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	high := make(chan string) // открыт, но пуст
	low := chkFilled("a", "", "b")
	out := Prioritize(ctx, high, low, 2)
	for _, want := range []string{"a", "", "b"} {
		select {
		case v := <-out:
			if v != want {
				t.Fatalf("получили %q, ожидали %q", v, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("high пуст, а значение %q из low не пришло", want)
		}
	}
	cancel()
	if got := chkDrain(t, out); len(got) != 0 {
		t.Fatalf("после отмены пришли значения %v", got)
	}
}

func TestPrioritizeOneClosed(t *testing.T) {
	high := make(chan string)
	close(high)
	low := make(chan string)
	go func() {
		for _, v := range []string{"x", "y", "z"} {
			low <- v
			time.Sleep(time.Millisecond)
		}
		close(low)
	}()
	got := chkDrain(t, Prioritize(context.Background(), high, low, 1))
	if !reflect.DeepEqual(got, []string{"x", "y", "z"}) {
		t.Fatalf("high закрыт сразу: получили %v, ожидали [x y z]", got)
	}
}

func TestPrioritizeCancelWhileSending(t *testing.T) {
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	out := Prioritize(ctx, chkFilled("h"), make(chan string), 1)
	time.Sleep(10 * time.Millisecond) // значение взято, но выход никто не читает
	cancel()
	chkDrain(t, out)
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после отмены осталось %d лишних горутин", n-before)
	}
}
