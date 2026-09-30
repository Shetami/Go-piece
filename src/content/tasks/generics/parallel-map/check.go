package main

var chkErrBoom = errors.New("boom")

func chkRun[U any](t *testing.T, call func() ([]U, error)) ([]U, error) {
	t.Helper()
	type res struct {
		out []U
		err error
	}
	ch := make(chan res, 1)
	go func() { out, err := call(); ch <- res{out, err} }()
	select {
	case r := <-ch:
		return r.out, r.err
	case <-time.After(5 * time.Second):
		t.Fatalf("ParallelMap не вернулся за 5 секунд — дедлок или ожидание отменённых задач")
		return nil, nil
	}
}

func TestParallelMapOrderAndLimit(t *testing.T) {
	var cur, peak atomic.Int32
	in := make([]int, 30)
	for i := range in {
		in[i] = i
	}
	out, err := chkRun(t, func() ([]string, error) {
		return ParallelMap(context.Background(), in, 3, func(ctx context.Context, x int) (string, error) {
			n := cur.Add(1)
			for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
			}
			time.Sleep(time.Duration(30-x) * time.Millisecond)
			cur.Add(-1)
			return strconv.Itoa(x * 10), nil
		})
	})
	if err != nil || len(out) != 30 || out[0] != "0" || out[29] != "290" || out[7] != "70" {
		t.Fatalf("out=%v err=%v — ожидали результаты по индексам входа", out, err)
	}
	if p := peak.Load(); p > 3 || p < 2 {
		t.Fatalf("одновременно работало до %d вызовов f, ожидали не больше limit=3 (и больше одного)", p)
	}
}

func TestParallelMapZeroLimit(t *testing.T) {
	var cur, peak atomic.Int32
	out, err := chkRun(t, func() ([]int, error) {
		return ParallelMap(context.Background(), []int{1, 2, 3, 4}, 0, func(ctx context.Context, x int) (int, error) {
			if n := cur.Add(1); n > peak.Load() {
				peak.Store(n)
			}
			time.Sleep(time.Millisecond)
			cur.Add(-1)
			return x * x, nil
		})
	})
	if err != nil || !reflect.DeepEqual(out, []int{1, 4, 9, 16}) || peak.Load() != 1 {
		t.Fatalf("limit=0: out=%v err=%v, одновременно %d — ожидали [1 4 9 16] и по одному", out, err, peak.Load())
	}
}

func TestParallelMapFirstError(t *testing.T) {
	var calls atomic.Int32
	in := make([]int, 100)
	for i := range in {
		in[i] = i
	}
	start := time.Now()
	out, err := chkRun(t, func() ([]int, error) {
		return ParallelMap(context.Background(), in, 4, func(ctx context.Context, x int) (int, error) {
			calls.Add(1)
			if x == 2 {
				return 0, chkErrBoom
			}
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(time.Second):
				return x, nil
			}
		})
	})
	if !errors.Is(err, chkErrBoom) || out != nil {
		t.Fatalf("out=%v err=%v — ожидали nil и ошибку упавшего f, а не context.Canceled от остальных", out, err)
	}
	if c := calls.Load(); c > 10 {
		t.Fatalf("после ошибки продолжили запускать f: всего %d вызовов из 100", c)
	}
	if d := time.Since(start); d > 900*time.Millisecond {
		t.Fatalf("ParallelMap работал %v — контекст для остальных f не отменили", d)
	}
}

func TestParallelMapOuterCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	out, err := chkRun(t, func() ([]int, error) {
		return ParallelMap(ctx, make([]int, 50), 2, func(ctx context.Context, x int) (int, error) {
			calls.Add(1)
			<-ctx.Done()
			return 0, ctx.Err()
		})
	})
	if !errors.Is(err, context.Canceled) || out != nil || calls.Load() > 4 {
		t.Fatalf("внешняя отмена: out=%v err=%v, вызовов f %d — ожидали nil, context.Canceled и не больше 4", out, err, calls.Load())
	}
}

func TestParallelMapNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 5 {
		chkRun(t, func() ([]int, error) {
			return ParallelMap(context.Background(), []int{1, 2, 3, 4, 5, 6}, 3, func(ctx context.Context, x int) (int, error) {
				if x == 1 {
					return 0, chkErrBoom
				}
				<-ctx.Done()
				return 0, ctx.Err()
			})
		})
	}
	time.Sleep(50 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("горутин было %d, стало %d — кто-то остался висеть после возврата", before, after)
	}
}
