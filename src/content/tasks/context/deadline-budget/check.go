package main

// chkRec — вызов, который запоминает, сколько времени ему дали, и сразу выходит.
func chkRec(name string, got map[string]time.Duration, ctxs *[]context.Context) Call {
	return Call{Name: name, Do: func(ctx context.Context) error {
		*ctxs = append(*ctxs, ctx)
		if dl, ok := ctx.Deadline(); ok {
			got[name] = time.Until(dl)
		} else {
			got[name] = -1
		}
		return nil
	}}
}

func chkBetween(t *testing.T, name string, d, lo, hi time.Duration) {
	t.Helper()
	if d < lo || d > hi {
		t.Fatalf("вызову %s дали %v, ожидали от %v до %v", name, d, lo, hi)
	}
}

func TestBudgetSplitsRemaining(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	defer cancel()
	got := map[string]time.Duration{}
	var ctxs []context.Context
	err := CallSequence(ctx, time.Millisecond, []Call{chkRec("a", got, &ctxs), chkRec("b", got, &ctxs), chkRec("c", got, &ctxs)})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	chkBetween(t, "a", got["a"], 200*time.Millisecond, 301*time.Millisecond)
	// a вернулся мгновенно — его сэкономленное время делят b и c.
	chkBetween(t, "b", got["b"], 380*time.Millisecond, 451*time.Millisecond)
	chkBetween(t, "c", got["c"], 780*time.Millisecond, 901*time.Millisecond)
	for i, c := range ctxs {
		if c.Err() == nil {
			t.Fatalf("контекст вызова %d жив после CallSequence — cancel не вызван", i+1)
		}
	}
}

func TestBudgetSlowCallTimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	var bCalled bool
	start := time.Now()
	err := CallSequence(ctx, time.Millisecond, []Call{
		{Name: "slow", Do: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }},
		{Name: "b", Do: func(context.Context) error { bCalled = true; return nil }},
	})
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "slow") {
		t.Fatalf("err = %v; ожидали DeadlineExceeded с именем slow", err)
	}
	if el := time.Since(start); el > 250*time.Millisecond {
		t.Fatalf("slow проработал %v — ему дали весь бюджет, а не половину", el)
	}
	if bCalled {
		t.Fatalf("после ошибки slow вызвали b")
	}
}

func TestBudgetTooSmall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	called := 0
	inc := Call{Name: "x", Do: func(context.Context) error { called++; return nil }}
	err := CallSequence(ctx, 50*time.Millisecond, []Call{inc, inc, {Name: "last", Do: inc.Do}})
	if !errors.Is(err, ErrNoBudget) || called != 0 {
		t.Fatalf("err = %v, вызовов %d; на троих 100 мс при минимуме 50 — ErrNoBudget сразу, без вызовов", err, called)
	}
}

func TestBudgetNoDeadline(t *testing.T) {
	got := map[string]time.Duration{}
	var ctxs []context.Context
	if err := CallSequence(context.Background(), time.Second, []Call{chkRec("a", got, &ctxs)}); err != nil {
		t.Fatalf("без дедлайна err = %v", err)
	}
	if got["a"] != -1 {
		t.Fatalf("у ctx нет дедлайна, а вызову выдали %v", got["a"])
	}
}

func TestBudgetOwnErrorAndCancel(t *testing.T) {
	bad := errors.New("404")
	err := CallSequence(context.Background(), 0, []Call{{Name: "user", Do: func(context.Context) error { return bad }}})
	if !errors.Is(err, bad) || !strings.Contains(err.Error(), "user") {
		t.Fatalf("err = %v; ожидали ошибку вызова с его именем", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err = CallSequence(ctx, 0, []Call{{Name: "a", Do: func(context.Context) error { called = true; return nil }}})
	if called || !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый ctx: вызван = %v, err = %v; ожидали без вызова и Canceled", called, err)
	}
}
