package main

type chkLog struct {
	mu  sync.Mutex
	log []string
}

func (l *chkLog) add(s string) { l.mu.Lock(); l.log = append(l.log, s); l.mu.Unlock() }
func (l *chkLog) get() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.log)
}

func chkRunShutdown(t *testing.T, s *Shutdown, ctx context.Context) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- s.Run(ctx, 50*time.Millisecond) }()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("Run не вернулся за 3 секунды — ждёт зависший хук?")
		return nil
	}
}

func TestShutdownOrder(t *testing.T) {
	var s Shutdown
	var l chkLog
	for _, name := range []string{"db", "cache", "http"} {
		s.Add(name, func(ctx context.Context) error { l.add(name); return nil })
	}
	if err := chkRunShutdown(t, &s, context.Background()); err != nil {
		t.Fatalf("Run = %v, ожидали nil", err)
	}
	if got := l.get(); !reflect.DeepEqual(got, []string{"http", "cache", "db"}) {
		t.Fatalf("порядок %v, ожидали [http cache db]", got)
	}
	if err := chkRunShutdown(t, &s, context.Background()); err != nil || len(l.get()) != 3 {
		t.Fatalf("повторный Run не должен ничего запускать")
	}
}

func TestShutdownHangingHook(t *testing.T) {
	var s Shutdown
	var l chkLog
	release := make(chan struct{})
	defer close(release)
	s.Add("db", func(ctx context.Context) error { l.add("db"); return nil })
	s.Add("cache", func(ctx context.Context) error { <-release; return nil }) // не смотрит на ctx
	err := chkRunShutdown(t, &s, context.Background())
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "cache") {
		t.Fatalf("Run = %v, ожидали \"cache: context deadline exceeded\"", err)
	}
	if got := l.get(); !reflect.DeepEqual(got, []string{"db"}) {
		t.Fatalf("после зависшего хука остальные должны выполниться, журнал %v", got)
	}
}

func TestShutdownCtxPerHook(t *testing.T) {
	var s Shutdown
	var gotDeadline atomic.Bool
	s.Add("flush", func(ctx context.Context) error {
		_, ok := ctx.Deadline()
		gotDeadline.Store(ok)
		<-ctx.Done()
		return ctx.Err()
	})
	err := chkRunShutdown(t, &s, context.Background())
	if !gotDeadline.Load() || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("хук должен получить контекст с таймаутом perHook: deadline=%v, err=%v", gotDeadline.Load(), err)
	}
}

func TestShutdownPanic(t *testing.T) {
	var s Shutdown
	var l chkLog
	errQ := errors.New("очередь не сбросилась")
	s.Add("db", func(ctx context.Context) error { l.add("db"); return nil })
	s.Add("queue", func(ctx context.Context) error { return errQ })
	s.Add("metrics", func(ctx context.Context) error { panic("nil exporter") })
	err := chkRunShutdown(t, &s, context.Background())
	if !strings.Contains(fmt.Sprint(err), "metrics") || !strings.Contains(fmt.Sprint(err), "nil exporter") {
		t.Fatalf("Run = %v, ожидали ошибку \"metrics: ... nil exporter\"", err)
	}
	if !errors.Is(err, errQ) || !strings.Contains(err.Error(), "queue") {
		t.Fatalf("Run = %v, ожидали и ошибку хука queue", err)
	}
	if !reflect.DeepEqual(l.get(), []string{"db"}) {
		t.Fatalf("после паники остальные хуки должны выполниться")
	}
}

func TestShutdownParentCanceled(t *testing.T) {
	var s Shutdown
	var l chkLog
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Add("db", func(ctx context.Context) error { l.add("db"); return nil })
	s.Add("http", func(ctx context.Context) error { l.add("http"); cancel(); return nil })
	err := chkRunShutdown(t, &s, ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run = %v, ожидали context.Canceled", err)
	}
	if got := l.get(); !reflect.DeepEqual(got, []string{"http"}) {
		t.Fatalf("после отмены ctx хуки не запускаются, журнал %v", got)
	}
}
