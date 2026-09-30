package main

func chkNoLeak(t *testing.T, base int) {
	t.Helper()
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine() - base; n > 0 {
		t.Fatalf("после возврата Run осталось %d лишних горутин", n)
	}
}

func TestRunOrderAndNoStages(t *testing.T) {
	inc := func(_ context.Context, v int) (int, error) { return v + 1, nil }
	dbl := func(_ context.Context, v int) (int, error) { return v * 2, nil }
	got, err := Run(context.Background(), []int{1, 2, 3, 4}, inc, dbl)
	if err != nil || !reflect.DeepEqual(got, []int{4, 6, 8, 10}) {
		t.Fatalf("Run(inc, dbl) = %v, %v; ожидали [4 6 8 10], nil", got, err)
	}
	src := []int{7, 8}
	got, err = Run(context.Background(), src)
	if err != nil || !reflect.DeepEqual(got, []int{7, 8}) {
		t.Fatalf("Run без этапов = %v, %v; ожидали [7 8], nil", got, err)
	}
}

func TestRunIsPipelined(t *testing.T) {
	slow := func(_ context.Context, v int) (int, error) {
		time.Sleep(40 * time.Millisecond)
		return v, nil
	}
	src := []int{1, 2, 3, 4, 5, 6, 7, 8}
	start := time.Now()
	got, _ := Run(context.Background(), src, slow, slow, slow)
	el := time.Since(start)
	if !reflect.DeepEqual(got, src) {
		t.Fatalf("Run = %v, ожидали %v", got, src)
	}
	// Последовательно: 8×3×40 = 960 мс; конвейером: (8+2)×40 = 400 мс.
	if el > 750*time.Millisecond {
		t.Fatalf("ушло %v — этапы не работают одновременно", el)
	}
}

func TestRunFirstErrorStopsAll(t *testing.T) {
	base := runtime.NumGoroutine()
	boom := errors.New("битая запись")
	var afterErr atomic.Int32
	failing := func(ctx context.Context, v int) (int, error) {
		if v == 3 {
			return 0, boom
		}
		return v, nil
	}
	// Второй этап медленный и вежливый: на отмене возвращает ctx.Err().
	polite := func(ctx context.Context, v int) (int, error) {
		select {
		case <-time.After(30 * time.Millisecond):
			return v, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
	counting := func(ctx context.Context, v int) (int, error) {
		if v > 3 {
			afterErr.Add(1)
		}
		return v, nil
	}
	src := make([]int, 1000)
	for i := range src {
		src[i] = i
	}
	got, err := Run(context.Background(), src, counting, failing, polite)
	if !errors.Is(err, boom) {
		t.Fatalf("Run вернул ошибку %v, ожидали ошибку этапа", err)
	}
	if got != nil {
		t.Fatalf("при ошибке ожидали nil-результат, получили %d значений", len(got))
	}
	if n := afterErr.Load(); n > 10 {
		t.Fatalf("после ошибки первый этап обработал ещё %d значений — конвейер не остановлен", n)
	}
	chkNoLeak(t, base)
}

func TestRunCancel(t *testing.T) {
	base := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	block := func(ctx context.Context, v int) (int, error) {
		time.Sleep(10 * time.Millisecond)
		return v, nil
	}
	done := make(chan error, 1)
	go func() {
		_, err := Run(ctx, make([]int, 10000), block, block)
		done <- err
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("после отмены Run вернул %v, ожидали context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("после отмены ctx Run не вернулся")
	}
	chkNoLeak(t, base)
}
