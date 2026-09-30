package main

var chkErrDB = errors.New("база недоступна")

type chkClock struct{ t time.Time }

func (c *chkClock) now() time.Time { return c.t }

func chkFail(context.Context) error { return chkErrDB }
func chkOK(context.Context) error   { return nil }

func TestBreakerOpens(t *testing.T) {
	c := &chkClock{time.Unix(1000, 0)}
	b := NewBreaker(3, time.Minute, c.now)
	for _, f := range []func(context.Context) error{chkFail, chkFail, chkOK, chkFail, chkFail} {
		b.Call(context.Background(), f)
	}
	calls := 0
	err := b.Call(context.Background(), func(context.Context) error { calls++; return chkErrDB })
	if calls != 1 || err != chkErrDB {
		t.Fatalf("успех сбрасывает счётчик: после fail,fail,ok,fail,fail цепь ещё замкнута; calls=%d err=%v", calls, err)
	}
	err = b.Call(context.Background(), func(context.Context) error { calls++; return nil })
	if calls != 1 {
		t.Fatal("после трёх сбоев подряд f вызываться не должна")
	}
	var oe *OpenError
	if !errors.Is(err, ErrOpen) || !errors.Is(err, chkErrDB) || !errors.As(err, &oe) {
		t.Fatalf("ожидали *OpenError, который Is(ErrOpen) и Is(причина), получили %v", err)
	}
	if !oe.Until.Equal(c.t.Add(time.Minute)) {
		t.Fatalf("Until = %v, ожидали now+cooldown = %v", oe.Until, c.t.Add(time.Minute))
	}
}

func TestBreakerIgnoresNotOurs(t *testing.T) {
	c := &chkClock{time.Unix(1000, 0)}
	b := NewBreaker(2, time.Minute, c.now)
	b.Call(context.Background(), chkFail)
	b.Call(context.Background(), func(context.Context) error { return fmt.Errorf("x: %w", context.Canceled) })
	b.Call(context.Background(), func(context.Context) error { return fmt.Errorf("нет email: %w", ErrClient) })
	calls := 0
	b.Call(context.Background(), func(context.Context) error { calls++; return chkErrDB })
	if calls != 1 {
		t.Fatal("Canceled и ErrClient не должны ни считаться сбоем, ни сбрасывать счётчик")
	}
	if err := b.Call(context.Background(), chkOK); !errors.Is(err, ErrOpen) {
		t.Fatalf("два настоящих сбоя подряд (с «чужими» ошибками между ними) размыкают цепь, а получили %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Call(ctx, chkOK); !errors.Is(err, context.Canceled) {
		t.Fatalf("отменённый ctx: ожидали context.Canceled, получили %v", err)
	}
}

func TestBreakerHalfOpen(t *testing.T) {
	c := &chkClock{time.Unix(1000, 0)}
	b := NewBreaker(1, time.Minute, c.now)
	b.Call(context.Background(), chkFail)
	c.t = c.t.Add(time.Minute)
	if err := b.Call(context.Background(), chkFail); err != chkErrDB {
		t.Fatalf("после cooldown пробный вызов должен выполниться, а получили %v", err)
	}
	if err := b.Call(context.Background(), chkOK); !errors.Is(err, ErrOpen) {
		t.Fatalf("неудачная проба снова размыкает цепь, а получили %v", err)
	}
	c.t = c.t.Add(time.Minute)
	b.Call(context.Background(), func(context.Context) error { return context.Canceled })
	if err := b.Call(context.Background(), chkOK); err != nil {
		t.Fatalf("проба с context.Canceled ничего не решает — следующий вызов тоже пробный, а получили %v", err)
	}
	for range 3 {
		if err := b.Call(context.Background(), chkOK); err != nil {
			t.Fatalf("после удачной пробы цепь замкнута, а получили %v", err)
		}
	}
}

func TestBreakerSingleProbe(t *testing.T) {
	c := &chkClock{time.Unix(1000, 0)}
	b := NewBreaker(1, time.Minute, c.now)
	b.Call(context.Background(), chkFail)
	c.t = c.t.Add(2 * time.Minute)

	release := make(chan struct{})
	var calls, rejected atomic.Int64
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := b.Call(context.Background(), func(context.Context) error {
				calls.Add(1)
				<-release
				return nil
			})
			if errors.Is(err, ErrOpen) {
				rejected.Add(1)
			}
		}()
	}
	deadline := time.After(2 * time.Second)
	for rejected.Load()+calls.Load() < 10 && calls.Load() <= 1 {
		select {
		case <-deadline:
			close(release)
			t.Fatalf("вызовы зависли: проба держит мьютекс? calls=%d rejected=%d", calls.Load(), rejected.Load())
		case <-time.After(time.Millisecond):
		}
	}
	close(release)
	wg.Wait()
	if calls.Load() != 1 || rejected.Load() != 9 {
		t.Fatalf("в полуоткрытом состоянии пробных вызовов %d, отклонено %d; ожидали 1 и 9", calls.Load(), rejected.Load())
	}
}
