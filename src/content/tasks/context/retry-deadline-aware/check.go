package main

var chkErrTemp = errors.New("503")

type chkRetryAfter struct{ d time.Duration }

func (e chkRetryAfter) Error() string             { return "429" }
func (e chkRetryAfter) RetryAfter() time.Duration { return e.d }

func TestDoBackoffSequence(t *testing.T) {
	var pauses []time.Duration
	calls := 0
	p := Policy{Attempts: 70, Base: time.Millisecond, Max: 50 * time.Millisecond,
		Jitter: func(d time.Duration) time.Duration { pauses = append(pauses, d); return 0 }}
	err := Do(context.Background(), p, func(context.Context) error { calls++; return chkErrTemp })
	if calls != 70 || !errors.Is(err, chkErrTemp) {
		t.Fatalf("calls = %d, err = %v; ожидали 70 попыток и последнюю ошибку", calls, err)
	}
	want := []time.Duration{1, 2, 4, 8, 16, 32, 50, 50}
	for i, w := range want {
		if pauses[i] != w*time.Millisecond {
			t.Fatalf("пауза перед попыткой %d = %v, ожидали %v (паузы: %v)", i+2, pauses[i], w*time.Millisecond, pauses[:8])
		}
	}
	if len(pauses) != 69 {
		t.Fatalf("Jitter вызван %d раз, ожидали 69 — после последней попытки паузы нет", len(pauses))
	}
	for i, d := range pauses {
		if d <= 0 || d > p.Max {
			t.Fatalf("пауза %d = %v — вышла за [0, Max]; переполнение при удвоении?", i+2, d)
		}
	}
}

func TestDoRetryAfter(t *testing.T) {
	jitterCalls, calls := 0, 0
	p := Policy{Attempts: 3, Base: time.Hour, Max: time.Hour,
		Jitter: func(d time.Duration) time.Duration { jitterCalls++; return d }}
	start := time.Now()
	err := Do(context.Background(), p, func(context.Context) error {
		calls++
		if calls == 1 {
			return fmt.Errorf("лимит: %w", chkRetryAfter{30 * time.Millisecond})
		}
		return nil
	})
	el := time.Since(start)
	if err != nil || calls != 2 {
		t.Fatalf("err = %v, calls = %d; ожидали успех со второй попытки", err, calls)
	}
	if el < 30*time.Millisecond || el > 10*time.Second {
		t.Fatalf("ждали %v; ожидали паузу из Retry-After (30 мс), а не Base", el)
	}
	if jitterCalls != 0 {
		t.Fatalf("к паузе из Retry-After применили Jitter")
	}
}

func TestDoNoTimeBeforeDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	calls := 0
	start := time.Now()
	err := Do(ctx, Policy{Attempts: 5, Base: time.Second, Max: time.Second}, func(context.Context) error {
		calls++
		return chkErrTemp
	})
	if time.Since(start) > 150*time.Millisecond {
		t.Fatalf("Do проспала %v, хотя пауза 1 с заведомо не влезала в дедлайн 200 мс", time.Since(start))
	}
	if calls != 1 || !errors.Is(err, ErrNoTime) || !errors.Is(err, chkErrTemp) {
		t.Fatalf("calls = %d, err = %v; ожидали 1 вызов и ошибку с ErrNoTime и последней ошибкой", calls, err)
	}
}

func TestDoCancelDuringPause(t *testing.T) {
	why := errors.New("остановка")
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(30 * time.Millisecond); cancel(why) }()
	calls := 0
	start := time.Now()
	err := Do(ctx, Policy{Attempts: 5, Base: time.Hour, Max: time.Hour}, func(context.Context) error {
		calls++
		return chkErrTemp
	})
	if time.Since(start) > 5*time.Second {
		t.Fatalf("отмена не прервала часовую паузу")
	}
	if calls != 1 || !errors.Is(err, why) || !errors.Is(err, chkErrTemp) {
		t.Fatalf("calls = %d, err = %v; ожидали причину отмены и последнюю ошибку f", calls, err)
	}
}

func TestDoPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := Do(ctx, Policy{Attempts: 3, Base: time.Millisecond, Max: time.Millisecond}, func(context.Context) error { called = true; return nil })
	if called || !errors.Is(err, context.Canceled) {
		t.Fatalf("called = %v, err = %v; ожидали, что f не вызывается, и Canceled", called, err)
	}
}
