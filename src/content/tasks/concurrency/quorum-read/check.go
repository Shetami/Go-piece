package main

// chkRep отвечает через d; hang — висит до отмены контекста.
func chkRep(r Reply, err error, d time.Duration, hang bool, canceled *atomic.Int32) Replica {
	return func(ctx context.Context) (Reply, error) {
		if hang {
			<-ctx.Done()
			canceled.Add(1)
			return Reply{}, ctx.Err()
		}
		time.Sleep(d)
		return r, err
	}
}

func chkQuorum(t *testing.T, ctx context.Context, reps []Replica, k int) (Reply, error) {
	t.Helper()
	type out struct {
		r   Reply
		err error
	}
	ch := make(chan out, 1)
	go func() { r, err := QuorumRead(ctx, reps, k); ch <- out{r, err} }()
	select {
	case o := <-ch:
		return o.r, o.err
	case <-time.After(2 * time.Second):
		t.Fatal("QuorumRead не вернулся — ждёт реплику, которая не ответит")
		return Reply{}, nil
	}
}

func TestQuorumFreshestOfK(t *testing.T) {
	var canceled atomic.Int32
	reps := []Replica{
		chkRep(Reply{"old", 1}, nil, 10*time.Millisecond, false, &canceled),
		chkRep(Reply{"new", 3}, nil, 30*time.Millisecond, false, &canceled),
		chkRep(Reply{}, nil, 0, true, &canceled), // зависшая
	}
	got, err := chkQuorum(t, context.Background(), reps, 2)
	if err != nil || got != (Reply{"new", 3}) {
		t.Fatalf("QuorumRead = %v, %v; ожидали {new 3} — самую свежую версию среди кворума", got, err)
	}
	for i := 0; i < 100 && canceled.Load() == 0; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if canceled.Load() != 1 {
		t.Fatal("зависшая реплика не получила отмену контекста после сбора кворума")
	}
}

func TestQuorumFailFast(t *testing.T) {
	var canceled atomic.Int32
	e1, e2 := errors.New("r1 down"), errors.New("r2 down")
	reps := []Replica{
		chkRep(Reply{}, e1, 10*time.Millisecond, false, &canceled),
		chkRep(Reply{}, e2, 20*time.Millisecond, false, &canceled),
		chkRep(Reply{}, nil, 0, true, &canceled),
	}
	_, err := chkQuorum(t, context.Background(), reps, 2)
	if !errors.Is(err, ErrNoQuorum) || !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("QuorumRead = %v; ожидали ошибку с ErrNoQuorum и обеими причинами", err)
	}
}

func TestQuorumToleratesFailures(t *testing.T) {
	var canceled atomic.Int32
	reps := []Replica{
		chkRep(Reply{}, errors.New("down"), 0, false, &canceled),
		chkRep(Reply{"a", 5}, nil, 20*time.Millisecond, false, &canceled),
		chkRep(Reply{"a", 5}, nil, 30*time.Millisecond, false, &canceled),
	}
	got, err := chkQuorum(t, context.Background(), reps, 2)
	if err != nil || got != (Reply{"a", 5}) {
		t.Fatalf("QuorumRead = %v, %v; одна упавшая реплика из трёх не мешает кворуму 2", got, err)
	}
}

func TestQuorumTooLarge(t *testing.T) {
	var calls atomic.Int32
	rep := func(ctx context.Context) (Reply, error) { calls.Add(1); return Reply{"x", 1}, nil }
	_, err := chkQuorum(t, context.Background(), []Replica{rep, rep}, 3)
	if !errors.Is(err, ErrNoQuorum) {
		t.Fatalf("k больше числа реплик: %v, ожидали ErrNoQuorum", err)
	}
	time.Sleep(10 * time.Millisecond)
	if calls.Load() != 0 {
		t.Fatal("при недостижимом кворуме реплики опрашивать не нужно")
	}
}

func TestQuorumCancelNoLeak(t *testing.T) {
	base := runtime.NumGoroutine()
	var canceled atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	slow := func(ctx context.Context) (Reply, error) {
		time.Sleep(80 * time.Millisecond) // не слушает контекст
		return Reply{"late", 1}, nil
	}
	reps := []Replica{slow, slow, chkRep(Reply{}, nil, 0, true, &canceled)}
	_, err := chkQuorum(t, ctx, reps, 2)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("QuorumRead = %v, ожидали DeadlineExceeded", err)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if d := runtime.NumGoroutine() - base; d > 0 {
		t.Fatalf("осталось %d горутин — опоздавшие реплики застряли на отправке результата", d)
	}
}
