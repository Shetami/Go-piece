package main

func checkDrain(t *testing.T, ch <-chan int) []int {
	t.Helper()
	if ch == nil {
		t.Fatal("Take вернул nil-канал")
	}
	var got []int
	timeout := time.After(2 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, v)
		case <-timeout:
			t.Fatalf("выходной канал не закрылся; успели прочитать %v", got)
		}
	}
}

func checkNaturals(ctx context.Context) <-chan int {
	ch := make(chan int)
	go func() {
		for i := 0; ; i++ {
			select {
			case ch <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func TestTakeFromInfinite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	got := checkDrain(t, Take(ctx, checkNaturals(ctx), 4))
	if !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
		t.Fatalf("Take 4 = %v, ожидали [0 1 2 3]", got)
	}
}

func TestTakeInputCloses(t *testing.T) {
	in := make(chan int, 2)
	in <- 7
	in <- 8
	close(in)
	got := checkDrain(t, Take(context.Background(), in, 10))
	if !reflect.DeepEqual(got, []int{7, 8}) {
		t.Fatalf("вход кончился раньше n: %v, ожидали [7 8]", got)
	}
}

func TestTakeCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int) // никто не пишет
	out := Take(ctx, in, 5)
	cancel()
	if got := checkDrain(t, out); len(got) != 0 {
		t.Fatalf("после отмены ничего не должно прийти: %v", got)
	}
}

func TestTakeZero(t *testing.T) {
	if got := checkDrain(t, Take(context.Background(), make(chan int), 0)); len(got) != 0 {
		t.Fatalf("Take 0 = %v", got)
	}
}
