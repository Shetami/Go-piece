package main

// chkReplica отвечает v через d (или падает с err), слушая ctx.
// started считает запуски, canceled — сколько раз её отменили.
type chkReplica struct {
	started, canceled atomic.Int32
}

func (r *chkReplica) make(d time.Duration, v string, err error) Replica {
	return func(ctx context.Context) (string, error) {
		r.started.Add(1)
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-t.C:
			return v, err
		case <-ctx.Done():
			r.canceled.Add(1)
			return "", ctx.Err()
		}
	}
}

func chkHedged(t *testing.T, ctx context.Context, delay time.Duration, rs []Replica) (string, error) {
	t.Helper()
	type out struct {
		v   string
		err error
	}
	ch := make(chan out, 1)
	go func() { v, err := Hedged(ctx, delay, rs); ch <- out{v, err} }()
	select {
	case o := <-ch:
		return o.v, o.err
	case <-time.After(3 * time.Second):
		t.Fatalf("Hedged не вернулся за 3 с")
		return "", nil
	}
}

func TestHedgedFastPrimary(t *testing.T) {
	var a, b chkReplica
	v, err := chkHedged(t, context.Background(), time.Hour, []Replica{a.make(time.Millisecond, "A", nil), b.make(0, "B", nil)})
	if v != "A" || err != nil || b.started.Load() != 0 {
		t.Fatalf("v = %q, err = %v, запусков B = %d; быстрый первый ответ — вторую реплику не трогаем", v, err, b.started.Load())
	}
}

func TestHedgedSlowPrimary(t *testing.T) {
	var a, b, c chkReplica
	v, err := chkHedged(t, context.Background(), 20*time.Millisecond, []Replica{
		a.make(time.Hour, "A", nil), b.make(5*time.Millisecond, "B", nil), c.make(0, "C", nil)})
	if v != "B" || err != nil {
		t.Fatalf("v = %q, err = %v; ожидали ответ второй реплики", v, err)
	}
	time.Sleep(50 * time.Millisecond)
	if a.canceled.Load() != 1 {
		t.Fatalf("медленная первая реплика не отменена после ответа второй")
	}
	if c.started.Load() != 0 {
		t.Fatalf("третья реплика запущена, хотя вторая ответила раньше её очереди")
	}
}

func TestHedgedErrorLaunchesNextNow(t *testing.T) {
	var a, b chkReplica
	start := time.Now()
	v, err := chkHedged(t, context.Background(), time.Hour, []Replica{
		a.make(0, "", errors.New("connection refused")), b.make(0, "B", nil)})
	if v != "B" || err != nil || time.Since(start) > time.Second {
		t.Fatalf("v = %q, err = %v за %v; упавшая реплика должна сразу запускать следующую", v, err, time.Since(start))
	}
}

func TestHedgedAllFail(t *testing.T) {
	e1, e2, e3 := errors.New("e1"), errors.New("e2"), errors.New("e3")
	var a, b, c chkReplica
	_, err := chkHedged(t, context.Background(), 10*time.Millisecond, []Replica{
		a.make(30*time.Millisecond, "", e1), b.make(0, "", e2), c.make(5*time.Millisecond, "", e3)})
	if !errors.Is(err, e1) || !errors.Is(err, e2) || !errors.Is(err, e3) {
		t.Fatalf("err = %v; ожидали все три ошибки", err)
	}
	if _, err := chkHedged(t, context.Background(), time.Millisecond, nil); !errors.Is(err, ErrNoReplicas) {
		t.Fatalf("без реплик err = %v, ожидали ErrNoReplicas", err)
	}
}

func TestHedgedParentCancel(t *testing.T) {
	why := errors.New("клиент ушёл")
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(30 * time.Millisecond); cancel(why) }()
	var a, b chkReplica
	_, err := chkHedged(t, ctx, 10*time.Millisecond, []Replica{a.make(time.Hour, "A", nil), b.make(time.Hour, "B", nil)})
	if !errors.Is(err, why) {
		t.Fatalf("err = %v, ожидали причину отмены", err)
	}
}

func TestHedgedNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 30 {
		var a, b, c chkReplica
		chkHedged(t, context.Background(), time.Millisecond, []Replica{
			a.make(time.Hour, "A", nil), b.make(time.Hour, "B", nil), c.make(2*time.Millisecond, "C", nil)})
	}
	time.Sleep(50 * time.Millisecond)
	if n := runtime.NumGoroutine(); n > before+5 {
		t.Fatalf("после 30 запросов горутин стало больше на %d — проигравшие реплики висят", n-before)
	}
}
