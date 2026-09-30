package main

func chkDone(t *testing.T, ctx context.Context, what string) {
	t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: контекст не отменился за 2 с", what)
	}
}

func TestLeaseRenewsWhileHealthy(t *testing.T) {
	var renews atomic.Int32
	ctx, cancel := WithLease(context.Background(), 10*time.Millisecond, func(context.Context) error { renews.Add(1); return nil })
	time.Sleep(105 * time.Millisecond)
	if ctx.Err() != nil {
		t.Fatalf("аренда продлевается успешно, а контекст отменён: %v", context.Cause(ctx))
	}
	if n := renews.Load(); n < 3 {
		t.Fatalf("за 10 интервалов продлений %d — renew не вызывается по расписанию", n)
	}
	cancel()
	cancel()
	if !errors.Is(context.Cause(ctx), context.Canceled) {
		t.Fatalf("после cancel Cause = %v, ожидали Canceled", context.Cause(ctx))
	}
}

func TestLeaseLostOnError(t *testing.T) {
	gone := errors.New("ключ в etcd перехвачен")
	var renews atomic.Int32
	ctx, cancel := WithLease(context.Background(), 10*time.Millisecond, func(context.Context) error {
		if renews.Add(1) == 3 {
			return gone
		}
		return nil
	})
	defer cancel()
	chkDone(t, ctx, "renew вернул ошибку")
	if c := context.Cause(ctx); !errors.Is(c, ErrLeaseLost) || !errors.Is(c, gone) {
		t.Fatalf("Cause = %v; ожидали ErrLeaseLost и ошибку renew", c)
	}
	time.Sleep(50 * time.Millisecond)
	if renews.Load() != 3 {
		t.Fatalf("после потери аренды продления продолжаются (%d)", renews.Load())
	}
}

func TestLeaseRenewTimeout(t *testing.T) {
	ctx, cancel := WithLease(context.Background(), 20*time.Millisecond, func(rctx context.Context) error {
		<-rctx.Done() // хранилище не отвечает
		return rctx.Err()
	})
	defer cancel()
	chkDone(t, ctx, "renew завис")
	if c := context.Cause(ctx); !errors.Is(c, ErrLeaseLost) || !errors.Is(c, context.DeadlineExceeded) {
		t.Fatalf("Cause = %v; зависшее продление должно кончиться таймаутом и потерей аренды", c)
	}
}

func TestLeaseParentCancel(t *testing.T) {
	why := errors.New("сервис останавливается")
	parent, cancelP := context.WithCancelCause(context.Background())
	before := runtime.NumGoroutine()
	ctx, _ := WithLease(parent, time.Hour, func(context.Context) error { return nil })
	cancelP(why)
	chkDone(t, ctx, "отмена родителя")
	if c := context.Cause(ctx); !errors.Is(c, why) {
		t.Fatalf("Cause = %v, ожидали причину родителя", c)
	}
	time.Sleep(20 * time.Millisecond)
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после отмены родителя фоновая горутина не вышла")
	}
}

func TestLeaseCancelWaitsForRenew(t *testing.T) {
	var inRenew atomic.Bool
	started := make(chan struct{}, 1)
	ctx, cancel := WithLease(context.Background(), 5*time.Millisecond, func(context.Context) error {
		inRenew.Store(true)
		select {
		case started <- struct{}{}:
		default:
		}
		time.Sleep(40 * time.Millisecond) // упрямое продление, контекст не слушает
		inRenew.Store(false)
		return nil
	})
	<-started
	cancel()
	if inRenew.Load() {
		t.Fatalf("cancel вернулся, пока renew ещё выполняется — работа под арендой может продолжиться")
	}
	if ctx.Err() == nil {
		t.Fatalf("после cancel контекст жив")
	}
}
