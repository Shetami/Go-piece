package main

type chkNode struct{ canceled atomic.Int32 }

// ok — реплика с ответом v через d; fail — с ошибкой; hang — ждёт отмены.
func (c *chkNode) ok(v int, d time.Duration) func(context.Context) (int, error) {
	return func(ctx context.Context) (int, error) {
		select {
		case <-time.After(d):
			return v, nil
		case <-ctx.Done():
			c.canceled.Add(1)
			return 0, ctx.Err()
		}
	}
}

func (c *chkNode) fail(err error) func(context.Context) (int, error) {
	return func(context.Context) (int, error) { return 0, err }
}

func (c *chkNode) hang() func(context.Context) (int, error) { return c.ok(0, time.Hour) }

func chkQuorum(t *testing.T, ctx context.Context, k int, rs []func(context.Context) (int, error)) ([]int, error) {
	t.Helper()
	type out struct {
		v   []int
		err error
	}
	ch := make(chan out, 1)
	go func() { v, err := Quorum(ctx, k, rs); ch <- out{v, err} }()
	select {
	case o := <-ch:
		return o.v, o.err
	case <-time.After(2 * time.Second):
		t.Fatalf("Quorum не вернулся за 2 с")
		return nil, nil
	}
}

func TestQuorumReached(t *testing.T) {
	var c chkNode
	got, err := chkQuorum(t, context.Background(), 3, []func(context.Context) (int, error){
		c.ok(1, time.Millisecond), c.hang(), c.ok(2, 2*time.Millisecond), c.hang(), c.ok(3, 3*time.Millisecond)})
	slices.Sort(got)
	if err != nil || !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("Quorum = %v, %v; ожидали [1 2 3] — ждать зависшие реплики не нужно", got, err)
	}
	time.Sleep(30 * time.Millisecond)
	if c.canceled.Load() != 2 {
		t.Fatalf("отменено %d лишних вызовов, ожидали 2", c.canceled.Load())
	}
}

func TestQuorumParallel(t *testing.T) {
	var c chkNode
	start := time.Now()
	rs := make([]func(context.Context) (int, error), 5)
	for i := range rs {
		rs[i] = c.ok(i, 100*time.Millisecond)
	}
	got, err := chkQuorum(t, context.Background(), 5, rs)
	if err != nil || len(got) != 5 || time.Since(start) > 400*time.Millisecond {
		t.Fatalf("5 реплик по 100 мс: %v, %v за %v — вызовы должны идти параллельно", got, err, time.Since(start))
	}
}

func TestQuorumFailsFast(t *testing.T) {
	var c chkNode
	e1, e2, e3 := errors.New("e1"), errors.New("e2"), errors.New("e3")
	_, err := chkQuorum(t, context.Background(), 3, []func(context.Context) (int, error){
		c.fail(e1), c.hang(), c.fail(e2), c.hang(), c.fail(e3)})
	if !errors.Is(err, ErrNoQuorum) || !errors.Is(err, e1) || !errors.Is(err, e2) || !errors.Is(err, e3) {
		t.Fatalf("err = %v; ожидали ErrNoQuorum и все три ошибки", err)
	}
	time.Sleep(30 * time.Millisecond)
	if c.canceled.Load() != 2 {
		t.Fatalf("зависшие реплики не отменены после провала кворума")
	}
}

func TestQuorumBadK(t *testing.T) {
	var calls atomic.Int32
	r := func(context.Context) (int, error) { calls.Add(1); return 1, nil }
	for _, k := range []int{0, 3} {
		if _, err := chkQuorum(t, context.Background(), k, []func(context.Context) (int, error){r, r}); !errors.Is(err, ErrNoQuorum) {
			t.Fatalf("k = %d из 2: err = %v, ожидали ErrNoQuorum", k, err)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("при недостижимом k реплики вызваны %d раз", calls.Load())
	}
}

func TestQuorumCancel(t *testing.T) {
	why := errors.New("дедлайн запроса")
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel(why) }()
	var c chkNode
	before := runtime.NumGoroutine()
	_, err := chkQuorum(t, ctx, 2, []func(context.Context) (int, error){c.ok(1, 0), c.hang(), c.hang()})
	if !errors.Is(err, why) {
		t.Fatalf("err = %v, ожидали причину отмены", err)
	}
	time.Sleep(30 * time.Millisecond)
	if n := runtime.NumGoroutine(); n > before+2 {
		t.Fatalf("после выхода висит %d лишних горутин", n-before)
	}
}
