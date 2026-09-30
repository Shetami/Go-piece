package main

type chkTemp struct {
	temp bool
	ra   time.Duration
}

func (e *chkTemp) Error() string   { return fmt.Sprintf("temp=%v", e.temp) }
func (e *chkTemp) Temporary() bool { return e.temp }

type chkBusy struct{ after time.Duration }

func (e chkBusy) Error() string             { return "сервер занят" }
func (e chkBusy) Temporary() bool           { return true }
func (e chkBusy) RetryAfter() time.Duration { return e.after }

type chkRec struct{ delays []time.Duration }

func (r *chkRec) sleep(ctx context.Context, d time.Duration) error {
	r.delays = append(r.delays, d)
	return ctx.Err()
}

func chkPolicy(n int, r *chkRec) Policy {
	return Policy{Attempts: n, Base: 10 * time.Millisecond, Max: 25 * time.Millisecond, Sleep: r.sleep}
}

func TestRetryBackoff(t *testing.T) {
	var r chkRec
	calls := 0
	err := Retry(context.Background(), chkPolicy(5, &r), func(context.Context) error {
		calls++
		if calls < 4 {
			return fmt.Errorf("запрос: %w", &chkTemp{temp: true})
		}
		return nil
	})
	want := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 25 * time.Millisecond}
	if err != nil || calls != 4 || !slices.Equal(r.delays, want) {
		t.Fatalf("err=%v calls=%d паузы=%v, ожидали nil, 4 вызова и паузы %v", err, calls, r.delays, want)
	}
}

func TestRetryPermanent(t *testing.T) {
	for _, perm := range []error{errors.New("неверный пароль"), fmt.Errorf("x: %w", &chkTemp{temp: false})} {
		var r chkRec
		calls := 0
		err := Retry(context.Background(), chkPolicy(5, &r), func(context.Context) error { calls++; return perm })
		if calls != 1 || err != perm || len(r.delays) != 0 {
			t.Fatalf("постоянная ошибка %q: calls=%d err=%v паузы=%v, ожидали 1 вызов, ту же ошибку и без пауз", perm, calls, err, r.delays)
		}
	}
}

func TestRetryExhausted(t *testing.T) {
	var r chkRec
	calls := 0
	last := &chkTemp{temp: true}
	err := Retry(context.Background(), chkPolicy(3, &r), func(context.Context) error { calls++; return last })
	if calls != 3 || len(r.delays) != 2 {
		t.Fatalf("calls=%d пауз=%d, ожидали 3 попытки и 2 паузы (после последней не ждут)", calls, len(r.delays))
	}
	if !errors.Is(err, last) || !strings.Contains(err.Error(), "3") {
		t.Fatalf("итоговая ошибка %q должна оборачивать последнюю и упоминать 3 попытки", err)
	}
}

func TestRetryAfterBeatsMax(t *testing.T) {
	var r chkRec
	calls := 0
	Retry(context.Background(), chkPolicy(2, &r), func(context.Context) error {
		calls++
		return fmt.Errorf("api: %w", chkBusy{after: time.Second})
	})
	if len(r.delays) != 1 || r.delays[0] != time.Second {
		t.Fatalf("паузы %v, ожидали [1s]: Retry-After от сервера важнее Max", r.delays)
	}
}

func TestRetryDeadlineIsTemporary(t *testing.T) {
	var r chkRec
	calls := 0
	err := Retry(context.Background(), chkPolicy(3, &r), func(context.Context) error {
		calls++
		if calls == 1 {
			return fmt.Errorf("попытка: %w", context.DeadlineExceeded) // таймаут одной попытки
		}
		return nil
	})
	if err != nil || calls != 2 {
		t.Fatalf("таймаут попытки при живом родительском ctx надо повторить: err=%v calls=%d", err, calls)
	}
}

func TestRetryContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := Retry(ctx, chkPolicy(3, &chkRec{}), func(context.Context) error { calls++; return nil })
	if calls != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый ctx: calls=%d err=%v, ожидали 0 вызовов и context.Canceled", calls, err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	last := &chkTemp{temp: true}
	calls = 0
	p := Policy{Attempts: 5, Base: time.Millisecond, Max: time.Second, Sleep: func(ctx context.Context, d time.Duration) error {
		cancel()
		return ctx.Err()
	}}
	err = Retry(ctx, p, func(context.Context) error { calls++; return last })
	if calls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, last) {
		t.Fatalf("отмена во время паузы: calls=%d err=%v, ожидали 1 вызов и ошибку с context.Canceled и последней ошибкой", calls, err)
	}
}

func TestRetryNoOverflow(t *testing.T) {
	var r chkRec
	p := Policy{Attempts: 80, Base: time.Second, Max: 10 * time.Second, Sleep: r.sleep}
	Retry(context.Background(), p, func(context.Context) error { return &chkTemp{temp: true} })
	for i, d := range r.delays {
		if d <= 0 || d > 10*time.Second {
			t.Fatalf("пауза №%d = %v, ожидали в (0, 10s] — переполнение Duration?", i+1, d)
		}
	}
}
