package main

func chkSource[T any](ctx context.Context, items ...T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, v := range items {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func chkDrain[T any](t *testing.T, ch <-chan T) []T {
	t.Helper()
	if ch == nil {
		t.Fatalf("этап вернул nil-канал")
	}
	var out []T
	timeout := time.After(3 * time.Second)
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-timeout:
			t.Fatalf("выходной канал не закрылся за 3 секунды; прочитано %v", out)
		}
	}
}

func TestPipelineCompose(t *testing.T) {
	ctx := context.Background()
	p := Then(Then(
		FilterStage(func(s string) bool { return s != "" }),
		MapStage(func(s string) int { return len([]rune(s)) })),
		BatchStage[int](2))
	got := chkDrain(t, p(ctx, chkSource(ctx, "go", "", "кот", "rust", "", "ёж")))
	want := [][]int{{2, 3}, {4, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("конвейер выдал %v, ожидали %v", got, want)
	}
}

func TestPipelineBatchTail(t *testing.T) {
	ctx := context.Background()
	got := chkDrain(t, BatchStage[int](3)(ctx, chkSource(ctx, 1, 2, 3, 4, 5, 6, 7)))
	if !reflect.DeepEqual(got, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}) {
		t.Fatalf("пачки по 3 из 7 элементов: %v", got)
	}
	got = chkDrain(t, BatchStage[int](2)(ctx, chkSource(ctx, 1, 2, 3, 4)))
	if !reflect.DeepEqual(got, [][]int{{1, 2}, {3, 4}}) {
		t.Fatalf("пачки по 2 из 4: %v — пустой хвост отправлять не нужно", got)
	}
	got[0][0] = 100
	if got[1][0] != 3 {
		t.Fatalf("пачки делят память: правка первой изменила вторую: %v", got)
	}
	if got := chkDrain(t, BatchStage[int](2)(ctx, chkSource[int](ctx))); len(got) != 0 {
		t.Fatalf("пустой вход дал пачки %v", got)
	}
}

func TestPipelineCancelNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 10 {
		ctx, cancel := context.WithCancel(context.Background())
		items := make([]int, 1000)
		p := Then(Then(MapStage(func(x int) int { return x + 1 }),
			FilterStage(func(x int) bool { return x > 0 })), BatchStage[int](4))
		out := p(ctx, chkSource(ctx, items...))
		if out == nil {
			t.Fatalf("этап вернул nil-канал")
		}
		<-out
		<-out
		cancel() // потребитель ушёл и больше не читает
	}
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before+2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("после отмены горутин было %d, стало %d — этапы висят на отправке в канал, который никто не читает", before, after)
	}
}
