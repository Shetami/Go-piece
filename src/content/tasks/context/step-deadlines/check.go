package main

// chkWork — шаг, который работает d и слушает контекст.
func chkWork(name string, d time.Duration, log *[]string) Step {
	return Step{Name: name, Run: func(ctx context.Context) error {
		*log = append(*log, name)
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-t.C:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}}
}

func TestPipelineOK(t *testing.T) {
	var log []string
	err := RunPipeline(context.Background(), time.Second, 100*time.Millisecond,
		[]Step{chkWork("a", 5*time.Millisecond, &log), chkWork("b", 5*time.Millisecond, &log)})
	if err != nil || !slices.Equal(log, []string{"a", "b"}) {
		t.Fatalf("err = %v, шаги %v; ожидали nil и [a b]", err, log)
	}
}

func TestPipelineStepTimeout(t *testing.T) {
	var log []string
	err := RunPipeline(context.Background(), 10*time.Second, 50*time.Millisecond, []Step{
		chkWork("a", time.Millisecond, &log), chkWork("slow", time.Hour, &log), chkWork("c", time.Millisecond, &log)})
	if !errors.Is(err, ErrStepTimeout) || !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrTotalTimeout) {
		t.Fatalf("err = %v; ожидали ErrStepTimeout и DeadlineExceeded (но не ErrTotalTimeout)", err)
	}
	if !strings.Contains(err.Error(), "slow") || !slices.Equal(log, []string{"a", "slow"}) {
		t.Fatalf("err = %v, шаги %v; ожидали имя шага в ошибке и остановку после slow", err, log)
	}
}

func TestPipelineTotalTimeout(t *testing.T) {
	var log []string
	// Каждый шаг укладывается в perStep, но вместе — нет: общий бюджет
	// должен оборвать шаг b посередине, а не только не пустить следующий.
	err := RunPipeline(context.Background(), 150*time.Millisecond, 100*time.Millisecond, []Step{
		chkWork("a", 80*time.Millisecond, &log), chkWork("b", 80*time.Millisecond, &log)})
	if !errors.Is(err, ErrTotalTimeout) || !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrStepTimeout) {
		t.Fatalf("err = %v; ожидали ErrTotalTimeout — общий бюджет 150 мс, два шага по 80", err)
	}
	if !strings.Contains(err.Error(), "b") {
		t.Fatalf("err = %v; ожидали, что общий таймаут прервёт шаг b", err)
	}
}

func TestPipelineParentCancel(t *testing.T) {
	why := errors.New("клиент ушёл")
	ctx, cancel := context.WithCancelCause(context.Background())
	var log []string
	err := RunPipeline(ctx, time.Hour, time.Hour, []Step{
		{Name: "a", Run: func(context.Context) error { cancel(why); return nil }},
		chkWork("b", time.Hour, &log)})
	if !errors.Is(err, why) || !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v; ожидали Canceled с причиной «клиент ушёл»", err)
	}
	if len(log) != 0 {
		t.Fatalf("после отмены запустился шаг b")
	}
}

func TestPipelineOwnError(t *testing.T) {
	bad := errors.New("нет такого пользователя")
	var log []string
	err := RunPipeline(context.Background(), time.Second, time.Second, []Step{
		{Name: "load", Run: func(context.Context) error { return bad }}, chkWork("b", 0, &log)})
	if !errors.Is(err, bad) || !strings.Contains(err.Error(), "load") || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v; ожидали исходную ошибку с именем шага", err)
	}
}

func TestPipelineReleasesStepCtx(t *testing.T) {
	var ctxs []context.Context
	step := Step{Name: "s", Run: func(ctx context.Context) error { ctxs = append(ctxs, ctx); return nil }}
	RunPipeline(context.Background(), time.Hour, time.Hour, []Step{step, step})
	for i, c := range ctxs {
		if c.Err() == nil {
			t.Fatalf("контекст шага %d жив после RunPipeline — cancel не вызван", i+1)
		}
	}
	if len(ctxs) == 2 && ctxs[0] == ctxs[1] {
		t.Fatalf("у шагов общий контекст — perStep не применяется к каждому шагу")
	}
}
