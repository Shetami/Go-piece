package main

type chkRes struct {
	v   string
	err error
}

func chkDo(g *Group[string, string], ctx context.Context, fn func(context.Context) (string, error)) chan chkRes {
	ch := make(chan chkRes, 1)
	go func() { v, err := g.Do(ctx, "user:1", fn); ch <- chkRes{v, err} }()
	return ch
}

func chkWait(t *testing.T, ch chan chkRes, what string) chkRes {
	t.Helper()
	select {
	case r := <-ch:
		return r
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: Do не вернулся за 2 с", what)
		return chkRes{}
	}
}

func TestGroupDedup(t *testing.T) {
	var g Group[string, string]
	var calls atomic.Int32
	gate := make(chan struct{})
	fn := func(context.Context) (string, error) { calls.Add(1); <-gate; return "alice", nil }
	var chans []chan chkRes
	for range 10 {
		chans = append(chans, chkDo(&g, context.Background(), fn))
	}
	time.Sleep(20 * time.Millisecond)
	close(gate)
	for _, ch := range chans {
		if r := chkWait(t, ch, "ждущий"); r.v != "alice" || r.err != nil {
			t.Fatalf("Do = %q, %v; ожидали alice", r.v, r.err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("fn вызвана %d раз на 10 одновременных Do, ожидали 1", calls.Load())
	}
	chkWait(t, chkDo(&g, context.Background(), fn), "следующий Do")
	if calls.Load() != 2 {
		t.Fatalf("после завершения вызова результат закэширован — новый Do должен вызвать fn")
	}
}

func TestGroupOneLeaves(t *testing.T) {
	var g Group[string, string]
	type key struct{}
	gate := make(chan struct{})
	fnErr := make(chan error, 1)
	fn := func(ctx context.Context) (string, error) {
		<-gate
		fnErr <- ctx.Err()
		return fmt.Sprint(ctx.Value(key{})), nil
	}
	why := errors.New("первый клиент ушёл")
	c1, cancel1 := context.WithCancelCause(context.WithValue(context.Background(), key{}, "req-1"))
	first := chkDo(&g, c1, fn)
	time.Sleep(10 * time.Millisecond)
	second := chkDo(&g, context.Background(), fn)
	time.Sleep(10 * time.Millisecond)
	cancel1(why)
	if r := chkWait(t, first, "отменённый"); !errors.Is(r.err, why) {
		t.Fatalf("отменённый вызывающий получил %v, ожидали причину своей отмены сразу", r.err)
	}
	close(gate)
	if err := <-fnErr; err != nil {
		t.Fatalf("контекст fn отменили, хотя второй вызывающий ещё ждёт: %v", err)
	}
	if r := chkWait(t, second, "второй"); r.v != "req-1" || r.err != nil {
		t.Fatalf("второй получил %q, %v; ожидали результат fn со значениями первого (req-1)", r.v, r.err)
	}
}

func TestGroupAllLeave(t *testing.T) {
	var g Group[string, string]
	var calls atomic.Int32
	abandoned := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	slow := func(ctx context.Context) (string, error) {
		calls.Add(1)
		<-ctx.Done()
		close(abandoned)
		<-release // брошенный вызов ещё не вернулся
		return "старый", nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	a, b := chkDo(&g, ctx, slow), chkDo(&g, ctx, slow)
	time.Sleep(10 * time.Millisecond)
	cancel()
	chkWait(t, a, "a")
	chkWait(t, b, "b")
	select {
	case <-abandoned:
	case <-time.After(2 * time.Second):
		t.Fatalf("все ждущие ушли, а контекст fn не отменён — работа идёт впустую")
	}
	r := chkWait(t, chkDo(&g, context.Background(), func(context.Context) (string, error) { calls.Add(1); return "новый", nil }), "новый Do")
	if r.v != "новый" || calls.Load() != 2 {
		t.Fatalf("новый Do получил %q (вызовов fn %d); он не должен присоединяться к брошенному вызову", r.v, calls.Load())
	}
}

func TestGroupErrorShared(t *testing.T) {
	var g Group[string, string]
	bad := errors.New("нет в базе")
	gate := make(chan struct{})
	fn := func(context.Context) (string, error) { <-gate; return "", bad }
	a, b := chkDo(&g, context.Background(), fn), chkDo(&g, context.Background(), fn)
	time.Sleep(10 * time.Millisecond)
	close(gate)
	if ra, rb := chkWait(t, a, "a"), chkWait(t, b, "b"); !errors.Is(ra.err, bad) || !errors.Is(rb.err, bad) {
		t.Fatalf("ошибки %v и %v; ожидали ошибку fn у обоих", ra.err, rb.err)
	}
}
