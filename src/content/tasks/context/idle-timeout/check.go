package main

func chkDoneIn(t *testing.T, ctx context.Context, what string) {
	t.Helper()
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatalf("%s: контекст не отменился за 2 с", what)
	}
}

func TestIdleSlidesWithActivity(t *testing.T) {
	// Запас большой: Sleep под нагрузкой затягивается, и при idle всего
	// втрое больше шага тест иногда падал на верном решении.
	ctx, touch, cancel := WithIdleTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	for range 40 { // 800 мс активности — вчетверо дольше idle
		time.Sleep(20 * time.Millisecond)
		touch()
		if ctx.Err() != nil {
			t.Fatalf("контекст отменён при регулярной активности: %v — touch должен сдвигать срок", context.Cause(ctx))
		}
	}
	start := time.Now()
	chkDoneIn(t, ctx, "активность прекратилась")
	if el := time.Since(start); el < 100*time.Millisecond {
		t.Fatalf("после последнего touch контекст прожил только %v, ожидали около 200 мс", el)
	}
	if !errors.Is(context.Cause(ctx), ErrIdle) || !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("Cause = %v, Err = %v; ожидали ErrIdle и Canceled", context.Cause(ctx), ctx.Err())
	}
}

func TestIdleNoTouch(t *testing.T) {
	ctx, _, cancel := WithIdleTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	chkDoneIn(t, ctx, "без активности")
	if !errors.Is(context.Cause(ctx), ErrIdle) {
		t.Fatalf("Cause = %v, ожидали ErrIdle", context.Cause(ctx))
	}
}

func TestIdleTouchAfterExpiry(t *testing.T) {
	ctx, touch, cancel := WithIdleTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	chkDoneIn(t, ctx, "простой")
	touch()
	touch()
	if ctx.Err() == nil || !errors.Is(context.Cause(ctx), ErrIdle) {
		t.Fatalf("touch после отмены изменил контекст: Err = %v, Cause = %v", ctx.Err(), context.Cause(ctx))
	}
}

func TestIdleParentAndCancel(t *testing.T) {
	why := errors.New("сервер останавливается")
	parent, cancelP := context.WithCancelCause(context.Background())
	ctx, _, cancel := WithIdleTimeout(parent, time.Hour)
	defer cancel()
	cancelP(why)
	chkDoneIn(t, ctx, "отмена родителя")
	if !errors.Is(context.Cause(ctx), why) {
		t.Fatalf("Cause = %v, ожидали причину родителя", context.Cause(ctx))
	}

	ctx2, touch2, cancel2 := WithIdleTimeout(context.Background(), 30*time.Millisecond)
	cancel2()
	cancel2()
	touch2()
	time.Sleep(60 * time.Millisecond)
	if !errors.Is(context.Cause(ctx2), context.Canceled) || errors.Is(context.Cause(ctx2), ErrIdle) {
		t.Fatalf("после cancel Cause = %v, ожидали Canceled (не ErrIdle)", context.Cause(ctx2))
	}
}

func TestIdleConcurrentTouch(t *testing.T) {
	ctx, touch, cancel := WithIdleTimeout(context.Background(), 200*time.Millisecond)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				touch()
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
	if ctx.Err() != nil {
		t.Fatalf("контекст отменён при непрерывной активности из многих горутин: %v", context.Cause(ctx))
	}
	cancel()
	touch()
}
