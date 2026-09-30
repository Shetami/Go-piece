package main

type chkFetch = func(context.Context) (string, error)

func chkOK(v string) chkFetch {
	return func(context.Context) (string, error) { return v, nil }
}

func chkFail(err error) chkFetch {
	return func(context.Context) (string, error) { return "", err }
}

// chkStuck висит до отмены своего контекста и отмечает, что отмену увидел.
func chkStuck(cancelled *atomic.Int64) chkFetch {
	return func(ctx context.Context) (string, error) {
		<-ctx.Done()
		cancelled.Add(1)
		return "", ctx.Err()
	}
}

func chkHedge(t *testing.T, ctx context.Context, delay time.Duration, fs ...chkFetch) (string, error) {
	t.Helper()
	type res struct {
		v   string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		v, err := Hedge(ctx, delay, fs...)
		ch <- res{v, err}
	}()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-time.After(2 * time.Second):
		t.Fatal("Hedge не вернулся за 2 секунды")
	}
	return "", nil
}

func TestHedgeFastFirst(t *testing.T) {
	var second atomic.Int64
	v, err := chkHedge(t, context.Background(), time.Hour, chkOK("a"),
		func(context.Context) (string, error) { second.Add(1); return "b", nil })
	if v != "a" || err != nil {
		t.Fatalf("Hedge = %q, %v; ожидали a, nil", v, err)
	}
	time.Sleep(10 * time.Millisecond)
	if second.Load() != 0 {
		t.Fatal("первая попытка ответила сразу, а вторая всё равно запущена")
	}
}

func TestHedgeSlowFirst(t *testing.T) {
	before := runtime.NumGoroutine()
	var cancelled atomic.Int64
	v, err := chkHedge(t, context.Background(), 20*time.Millisecond, chkStuck(&cancelled), chkOK("b"))
	if v != "b" || err != nil {
		t.Fatalf("Hedge = %q, %v; ожидали b от подстраховки", v, err)
	}
	deadline := time.Now().Add(time.Second)
	for (cancelled.Load() == 0 || runtime.NumGoroutine() > before) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if cancelled.Load() == 0 {
		t.Fatal("проигравшая попытка не получила отмену контекста")
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после Hedge висят %d горутин проигравших попыток", n-before)
	}
}

func TestHedgeFailureLaunchesNext(t *testing.T) {
	v, err := chkHedge(t, context.Background(), time.Hour, chkFail(errors.New("503")), chkOK("b"))
	if v != "b" || err != nil {
		t.Fatalf("Hedge = %q, %v; ожидали b: после ошибки следующая попытка — сразу, а не через delay", v, err)
	}
}

func TestHedgeAllFail(t *testing.T) {
	e1, e2, e3 := errors.New("e1"), errors.New("e2"), errors.New("e3")
	_, err := chkHedge(t, context.Background(), time.Hour, chkFail(e1), chkFail(e2), chkFail(e3))
	for _, e := range []error{e1, e2, e3} {
		if !errors.Is(err, e) {
			t.Fatalf("все упали: ошибка %v не содержит %v", err, e)
		}
	}
	if err.Error() != "e1\ne2\ne3" {
		t.Fatalf("ошибка %q, ожидали errors.Join в порядке fetchers: \"e1\\ne2\\ne3\"", err.Error())
	}
	if _, err := chkHedge(t, context.Background(), time.Hour); !errors.Is(err, ErrNoFetchers) {
		t.Fatalf("без fetchers = %v, ожидали ErrNoFetchers", err)
	}
}

func TestHedgeParentCancel(t *testing.T) {
	var cancelled atomic.Int64
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := chkHedge(t, ctx, 5*time.Millisecond, chkStuck(&cancelled), chkStuck(&cancelled))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("по дедлайну родителя Hedge = %v, ожидали DeadlineExceeded", err)
	}
}
