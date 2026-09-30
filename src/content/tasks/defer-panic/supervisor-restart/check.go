package main

func chkCrashingWorker(ctx context.Context) error { panic("очередь вернула мусор") }

func chkSupervise(t *testing.T, f func() error) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- f() }()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("Supervise не вернулся за 3 секунды")
		return nil
	}
}

func chkBackoff(n int) time.Duration { return time.Millisecond }

func TestSuperviseRestartsThenOK(t *testing.T) {
	runs := 0
	var seen []int
	err := chkSupervise(t, func() error {
		return Supervise(context.Background(), 5, chkBackoff,
			func(n int, pe *PanicError) { seen = append(seen, n) },
			func(ctx context.Context) error {
				runs++
				if runs <= 2 {
					panic(fmt.Sprintf("сбой %d", runs))
				}
				return nil
			})
	})
	if err != nil || runs != 3 {
		t.Fatalf("Supervise = %v после %d запусков; ожидали nil после 3", err, runs)
	}
	if !reflect.DeepEqual(seen, []int{1, 2}) {
		t.Fatalf("onPanic вызван с %v, ожидали [1 2]", seen)
	}
}

func TestSuperviseErrorNoRestart(t *testing.T) {
	runs := 0
	fatal := errors.New("конфиг невалиден")
	err := chkSupervise(t, func() error {
		return Supervise(context.Background(), 5, chkBackoff, func(int, *PanicError) {},
			func(ctx context.Context) error { runs++; return fatal })
	})
	if err != fatal || runs != 1 {
		t.Fatalf("Supervise = %v после %d запусков; ошибку не перезапускают", err, runs)
	}
}

func TestSuperviseGivesUp(t *testing.T) {
	runs := 0
	var last *PanicError
	err := chkSupervise(t, func() error {
		return Supervise(context.Background(), 2, chkBackoff,
			func(n int, pe *PanicError) { last = pe },
			func(ctx context.Context) error { runs++; return chkCrashingWorker(ctx) })
	})
	if runs != 3 {
		t.Fatalf("запусков %d, ожидали 3: первый и два перезапуска", runs)
	}
	var pe *PanicError
	if !errors.As(err, &pe) || pe.Value != "очередь вернула мусор" {
		t.Fatalf("Supervise = %v, ожидали ошибку с *PanicError последней паники", err)
	}
	if last == nil || !strings.Contains(string(last.Stack), "chkCrashingWorker") {
		t.Fatalf("в стеке PanicError должна быть паникующая функция chkCrashingWorker")
	}
}

func TestSuperviseCancelDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runs := 0
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	err := chkSupervise(t, func() error {
		return Supervise(ctx, 10, func(int) time.Duration { return time.Hour }, func(int, *PanicError) {},
			func(ctx context.Context) error { runs++; panic("x") })
	})
	if !errors.Is(err, context.Canceled) || runs != 1 {
		t.Fatalf("Supervise = %v после %d запусков; пауза должна прерываться отменой ctx", err, runs)
	}
}

func TestSuperviseCanceledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runs := 0
	err := Supervise(ctx, 3, chkBackoff, func(int, *PanicError) {}, func(context.Context) error { runs++; return nil })
	if !errors.Is(err, context.Canceled) || runs != 0 {
		t.Fatalf("ctx уже отменён: Supervise = %v, запусков %d; ожидали Canceled и 0", err, runs)
	}
}
