package main

var chkErrTemp = errors.New("503 от апстрима")

type chkWaiter struct{ calls []int }

func (w *chkWaiter) wait(ctx context.Context, n int) error {
	w.calls = append(w.calls, n)
	return ctx.Err()
}

func TestRetryEventuallySucceeds(t *testing.T) {
	var w chkWaiter
	calls := 0
	err := Retry(context.Background(), 5, w.wait, func(context.Context) error {
		calls++
		if calls < 3 {
			return chkErrTemp
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("Retry = %v, вызовов %d; ожидали nil и 3", err, calls)
	}
	if !reflect.DeepEqual(w.calls, []int{1, 2}) {
		t.Fatalf("wait вызван с %v, ожидали [1 2]", w.calls)
	}
}

func TestRetryExhausted(t *testing.T) {
	var w chkWaiter
	calls := 0
	err := Retry(context.Background(), 3, w.wait, func(context.Context) error {
		calls++
		return fmt.Errorf("попытка %d: %w", calls, chkErrTemp)
	})
	if calls != 3 || !errors.Is(err, chkErrTemp) || !strings.Contains(err.Error(), "попытка 3") {
		t.Fatalf("Retry = %v после %d вызовов; ожидали ошибку третьей попытки", err, calls)
	}
	if !reflect.DeepEqual(w.calls, []int{1, 2}) {
		t.Fatalf("wait вызван с %v, ожидали [1 2] — после последней попытки ждать незачем", w.calls)
	}
}

func TestRetryPanicNotRetried(t *testing.T) {
	var w chkWaiter
	calls := 0
	bad := errors.New("инвариант нарушен")
	err := Retry(context.Background(), 5, w.wait, func(context.Context) error {
		calls++
		if calls == 2 {
			panic(bad)
		}
		return chkErrTemp
	})
	var pe *PanicError
	if !errors.As(err, &pe) || !errors.Is(err, bad) {
		t.Fatalf("Retry = %v, ожидали *PanicError, через который виден исходный error", err)
	}
	if calls != 2 || len(w.calls) != 1 {
		t.Fatalf("после паники повторов быть не должно: вызовов %d, ожиданий %d", calls, len(w.calls))
	}
}

func TestRetryStringPanic(t *testing.T) {
	err := Retry(context.Background(), 3, (&chkWaiter{}).wait, func(context.Context) error { panic("индекс вне диапазона") })
	var pe *PanicError
	if !errors.As(err, &pe) || pe.Value != "индекс вне диапазона" {
		t.Fatalf("Retry = %v, ожидали *PanicError со значением паники", err)
	}
}

func TestRetryWaitCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	err := Retry(ctx, 10, func(ctx context.Context, n int) error {
		cancel()
		return ctx.Err()
	}, func(context.Context) error { calls++; return chkErrTemp })
	if calls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, chkErrTemp) {
		t.Fatalf("Retry = %v после %d вызовов; ожидали 1 вызов и обе ошибки", err, calls)
	}
}

func TestRetryZeroAttempts(t *testing.T) {
	calls := 0
	Retry(context.Background(), 0, (&chkWaiter{}).wait, func(context.Context) error { calls++; return chkErrTemp })
	if calls != 1 {
		t.Fatalf("attempts=0: вызовов %d, ожидали 1", calls)
	}
}
