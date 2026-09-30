package main

func chkLeft(t *testing.T, ctx context.Context) time.Duration {
	t.Helper()
	dl, ok := ctx.Deadline()
	if !ok {
		t.Fatalf("у контекста нет дедлайна")
	}
	return time.Until(dl)
}

func TestInject(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	h := map[string]string{}
	Inject(ctx, h, 100*time.Millisecond)
	ms, err := strconv.Atoi(h[TimeoutHeader])
	if err != nil || ms < 850 || ms > 900 {
		t.Fatalf("заголовок = %q; ожидали целые мс около 900 (1 с минус запас 100 мс, вниз)", h[TimeoutHeader])
	}
	short, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel2()
	Inject(short, h, 100*time.Millisecond)
	if h[TimeoutHeader] != "0" {
		t.Fatalf("остаток меньше запаса: заголовок = %q, ожидали \"0\"", h[TimeoutHeader])
	}
	Inject(context.Background(), h, 0)
	if _, ok := h[TimeoutHeader]; ok {
		t.Fatalf("у ctx нет дедлайна, а заголовок остался: %q", h[TimeoutHeader])
	}
}

func TestExtractValid(t *testing.T) {
	ctx, cancel := Extract(context.Background(), map[string]string{TimeoutHeader: "300"}, time.Minute)
	defer cancel()
	if d := chkLeft(t, ctx); d > 300*time.Millisecond || d < 250*time.Millisecond {
		t.Fatalf("заголовок 300: осталось %v, ожидали ~300 мс", d)
	}
	ctx2, cancel2 := Extract(context.Background(), map[string]string{TimeoutHeader: "600000"}, 2*time.Second)
	defer cancel2()
	if d := chkLeft(t, ctx2); d > 2*time.Second {
		t.Fatalf("клиент попросил 10 минут, а лимит 2 с: осталось %v", d)
	}
	ctx3, cancel3 := Extract(context.Background(), map[string]string{TimeoutHeader: "0"}, time.Minute)
	defer cancel3()
	if !errors.Is(ctx3.Err(), context.DeadlineExceeded) {
		t.Fatalf("заголовок \"0\": Err = %v, ожидали уже истёкший контекст", ctx3.Err())
	}
}

func TestExtractInvalid(t *testing.T) {
	for _, v := range []string{"abc", "-5", "1.5", "", " 100", "99999999999999999999"} {
		ctx, cancel := Extract(context.Background(), map[string]string{TimeoutHeader: v}, 500*time.Millisecond)
		d := chkLeft(t, ctx)
		cancel()
		if d < 400*time.Millisecond || d > 500*time.Millisecond {
			t.Fatalf("заголовок %q: осталось %v, ожидали лимит 500 мс", v, d)
		}
	}
	ctx, cancel := Extract(context.Background(), nil, 500*time.Millisecond)
	defer cancel()
	if d := chkLeft(t, ctx); d < 400*time.Millisecond {
		t.Fatalf("без заголовков: осталось %v, ожидали лимит 500 мс", d)
	}
}

func TestExtractHugeNoOverflow(t *testing.T) {
	// 9e15 мс корректно парсится, но ×1e6 не влезает в int64.
	ctx, cancel := Extract(context.Background(), map[string]string{TimeoutHeader: "9000000000000000"}, time.Second)
	defer cancel()
	if ctx.Err() != nil {
		t.Fatalf("огромный таймаут превратился в истёкший контекст — переполнение Duration")
	}
	if d := chkLeft(t, ctx); d > time.Second {
		t.Fatalf("осталось %v, ожидали не больше лимита", d)
	}
}

func TestExtractKeepsParent(t *testing.T) {
	type key struct{}
	parent, cancelP := context.WithTimeout(context.WithValue(context.Background(), key{}, "v"), 50*time.Millisecond)
	defer cancelP()
	ctx, cancel := Extract(parent, map[string]string{TimeoutHeader: "5000"}, time.Minute)
	defer cancel()
	if d := chkLeft(t, ctx); d > 50*time.Millisecond || ctx.Value(key{}) != "v" {
		t.Fatalf("осталось %v; дедлайн и значения родителя должны сохраниться", d)
	}
	cancel()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("возвращённая cancel не отменяет контекст")
	}
}
