package main

func chkWaitErr(t *testing.T, l *Limiter, ctx context.Context) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- l.Wait(ctx) }()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatalf("Wait не вернулся за 3 с")
		return nil
	}
}

func TestLimiterSpacing(t *testing.T) {
	l := NewLimiter(40 * time.Millisecond)
	start := time.Now()
	var passed [5]time.Duration
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Wait(context.Background()); err != nil {
				t.Errorf("Wait: %v", err)
			}
			passed[i] = time.Since(start)
		}()
	}
	wg.Wait()
	s := passed[:]
	slices.Sort(s)
	if s[0] > 30*time.Millisecond {
		t.Fatalf("первое событие ждало %v — оно должно пройти сразу", s[0])
	}
	for i, d := range s {
		if d < time.Duration(i)*40*time.Millisecond-time.Millisecond {
			t.Fatalf("событие №%d прошло через %v — раньше своего слота %v (все: %v)", i+1, d, time.Duration(i)*40*time.Millisecond, s)
		}
	}
}

func TestLimiterDeadlineTooSoon(t *testing.T) {
	l := NewLimiter(100 * time.Millisecond)
	start := time.Now()
	chkWaitErr(t, l, context.Background()) // занял слот t=0
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := chkWaitErr(t, l, ctx); !errors.Is(err, ErrDeadlineTooSoon) {
		t.Fatalf("Wait = %v; следующий слот через 100 мс, дедлайн через 10 — ожидали ErrDeadlineTooSoon", err)
	}
	if el := time.Since(start); el > 50*time.Millisecond {
		t.Fatalf("Wait с заведомо недостижимым слотом ждал %v вместо немедленного отказа", el)
	}
	chkWaitErr(t, l, context.Background())
	if el := time.Since(start); el > 170*time.Millisecond {
		t.Fatalf("следующий Wait прошёл через %v, ожидали ~100 мс — отказ не должен занимать слот", el)
	}
}

func TestLimiterCancelReturnsSlot(t *testing.T) {
	why := errors.New("клиент ушёл")
	l := NewLimiter(100 * time.Millisecond)
	start := time.Now()
	chkWaitErr(t, l, context.Background())
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel(why) }()
	if err := chkWaitErr(t, l, ctx); !errors.Is(err, why) {
		t.Fatalf("Wait = %v, ожидали причину отмены", err)
	}
	if el := time.Since(start); el > 70*time.Millisecond {
		t.Fatalf("отмена не прервала ожидание слота (прошло %v)", el)
	}
	chkWaitErr(t, l, context.Background())
	if el := time.Since(start); el > 170*time.Millisecond {
		t.Fatalf("следующий Wait прошёл через %v, ожидали ~100 мс — отменённый должен вернуть слот", el)
	}
}

func TestLimiterPreCanceled(t *testing.T) {
	l := NewLimiter(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := chkWaitErr(t, l, ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait с отменённым ctx = %v, ожидали Canceled", err)
	}
	start := time.Now()
	chkWaitErr(t, l, context.Background())
	if time.Since(start) > 30*time.Millisecond {
		t.Fatalf("отменённый заранее Wait занял слот — следующий ждёт")
	}
}
