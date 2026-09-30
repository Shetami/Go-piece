package main

type chkKey struct{}

func TestBackgroundOutlivesRequest(t *testing.T) {
	b := NewBackground(time.Second, 4)
	req, cancel := context.WithCancel(context.WithValue(context.Background(), chkKey{}, "req-42"))
	type seen struct {
		err    error
		val    any
		hasDl  bool
		dlLeft time.Duration
	}
	got := make(chan seen, 1)
	ok := b.Go(req, func(ctx context.Context) {
		time.Sleep(30 * time.Millisecond) // запрос к этому моменту уже завершён
		dl, has := ctx.Deadline()
		got <- seen{ctx.Err(), ctx.Value(chkKey{}), has, time.Until(dl)}
	})
	cancel()
	if !ok {
		t.Fatalf("Go вернул false при свободных местах")
	}
	s := <-got
	if s.err != nil {
		t.Fatalf("фоновая задача отменилась вместе с запросом: %v", s.err)
	}
	if s.val != "req-42" {
		t.Fatalf("значение из контекста запроса потеряно: %v", s.val)
	}
	if !s.hasDl || s.dlLeft > time.Second {
		t.Fatalf("у фоновой задачи нет своего дедлайна в 1 с (есть: %v, осталось %v)", s.hasDl, s.dlLeft)
	}
}

func TestBackgroundTimeout(t *testing.T) {
	b := NewBackground(30*time.Millisecond, 1)
	res := make(chan error, 1)
	b.Go(context.Background(), func(ctx context.Context) {
		select {
		case <-ctx.Done():
			res <- ctx.Err()
		case <-time.After(5 * time.Second):
			res <- nil
		}
	})
	if err := <-res; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("задача не получила свой таймаут: %v", err)
	}
}

func TestBackgroundLimit(t *testing.T) {
	b := NewBackground(time.Second, 2)
	gate := make(chan struct{})
	block := func(context.Context) { <-gate }
	if !b.Go(context.Background(), block) || !b.Go(context.Background(), block) {
		t.Fatalf("две задачи при limit=2 должны запуститься")
	}
	start := time.Now()
	if b.Go(context.Background(), block) {
		t.Fatalf("третья задача запустилась при limit=2")
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("Go ждал свободного места вместо того, чтобы сразу вернуть false")
	}
	close(gate)
	if err := b.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	b2 := NewBackground(time.Second, 1)
	b2.Go(context.Background(), func(context.Context) {})
	time.Sleep(20 * time.Millisecond)
	if !b2.Go(context.Background(), func(context.Context) {}) {
		t.Fatalf("место не освободилось после завершения задачи")
	}
	b2.Shutdown(context.Background())
}

func TestBackgroundShutdownWaits(t *testing.T) {
	b := NewBackground(time.Second, 10)
	var done atomic.Int32
	for range 5 {
		b.Go(context.Background(), func(context.Context) { time.Sleep(20 * time.Millisecond); done.Add(1) })
	}
	if err := b.Shutdown(context.Background()); err != nil || done.Load() != 5 {
		t.Fatalf("Shutdown = %v, завершилось %d из 5; ожидали дождаться всех", err, done.Load())
	}
	if b.Go(context.Background(), func(context.Context) {}) {
		t.Fatalf("после Shutdown Go должен возвращать false")
	}
	if err := b.Shutdown(context.Background()); err != nil {
		t.Fatalf("повторный Shutdown: %v", err)
	}
}

func TestBackgroundShutdownDeadline(t *testing.T) {
	b := NewBackground(time.Hour, 1)
	gate := make(chan struct{})
	defer close(gate)
	b.Go(context.Background(), func(context.Context) { <-gate })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	res := make(chan error, 1)
	go func() { res <- b.Shutdown(ctx) }()
	select {
	case err := <-res:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Shutdown = %v, ожидали DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Shutdown не уважает свой ctx и ждёт зависшую задачу")
	}
}
