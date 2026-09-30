package main

type chkLog struct{ lines []string }

// comp — компонент, который пишет в лог; start/stop можно подменить.
func (l *chkLog) comp(name string, start, stop func(context.Context) error) Component {
	return Component{
		Name: name,
		Start: func(ctx context.Context) error {
			l.lines = append(l.lines, "start "+name)
			if start != nil {
				return start(ctx)
			}
			return nil
		},
		Stop: func(ctx context.Context) error {
			l.lines = append(l.lines, fmt.Sprintf("stop %s (ctx жив: %v)", name, ctx.Err() == nil))
			if stop != nil {
				return stop(ctx)
			}
			return nil
		},
	}
}

func TestAppOrder(t *testing.T) {
	var l chkLog
	app := &App{StopTimeout: time.Second, Components: []Component{l.comp("db", nil, nil), l.comp("cache", nil, nil), l.comp("http", nil, nil)}}
	if err := app.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := app.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	want := []string{"start db", "start cache", "start http", "stop http (ctx жив: true)", "stop cache (ctx жив: true)", "stop db (ctx жив: true)"}
	if !slices.Equal(l.lines, want) {
		t.Fatalf("лог %q\nожидали %q", l.lines, want)
	}
	l.lines = nil
	if err := app.Stop(context.Background()); err != nil || len(l.lines) != 0 {
		t.Fatalf("повторный Stop: err = %v, вызовы %q; ожидали, что он ничего не делает", err, l.lines)
	}
}

func TestAppStartFailRollsBack(t *testing.T) {
	var l chkLog
	bad := errors.New("не удалось подключиться")
	app := &App{StopTimeout: time.Second, Components: []Component{
		l.comp("db", nil, nil), l.comp("cache", func(context.Context) error { return bad }, nil), l.comp("http", nil, nil)}}
	err := app.Start(context.Background())
	if !errors.Is(err, bad) || !strings.Contains(err.Error(), "start cache") {
		t.Fatalf("err = %v; ожидали ошибку старта с именем cache", err)
	}
	want := []string{"start db", "start cache", "stop db (ctx жив: true)"}
	if !slices.Equal(l.lines, want) {
		t.Fatalf("лог %q\nожидали %q — запущенное откатывается, упавший и незапущенные не останавливаются", l.lines, want)
	}
}

func TestAppStartCanceledRollbackGetsLiveCtx(t *testing.T) {
	var l chkLog
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{StopTimeout: time.Second, Components: []Component{
		l.comp("db", nil, nil),
		l.comp("consumer", func(ctx context.Context) error { cancel(); return ctx.Err() }, nil)}}
	err := app.Start(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, ожидали Canceled", err)
	}
	if len(l.lines) != 3 || l.lines[2] != "stop db (ctx жив: true)" {
		t.Fatalf("лог %q; откат после отмены старта должен получить живой контекст", l.lines)
	}
}

func TestAppStopErrorsAndTimeouts(t *testing.T) {
	var l chkLog
	flush := errors.New("не сбросили буфер")
	hang := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	app := &App{StopTimeout: 30 * time.Millisecond, Components: []Component{
		l.comp("db", nil, nil),
		l.comp("cache", nil, func(context.Context) error { return flush }),
		l.comp("consumer", nil, hang),
		l.comp("http", nil, hang)}}
	app.Start(context.Background())
	l.lines = nil
	done := make(chan error, 1)
	go func() { done <- app.Stop(context.Background()) }()
	var err error
	select {
	case err = <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("Stop висит — у зависших компонентов нет таймаута")
	}
	if !errors.Is(err, flush) || !errors.Is(err, context.DeadlineExceeded) ||
		!strings.Contains(err.Error(), "stop cache") || !strings.Contains(err.Error(), "stop consumer") || !strings.Contains(err.Error(), "stop http") {
		t.Fatalf("err = %v; ожидали ошибки всех трёх компонентов с именами", err)
	}
	want := []string{"stop http (ctx жив: true)", "stop consumer (ctx жив: true)", "stop cache (ctx жив: true)", "stop db (ctx жив: true)"}
	if !slices.Equal(l.lines, want) {
		t.Fatalf("лог %q\nожидали %q — каждый Stop со своим таймаутом, ошибки не прерывают остановку", l.lines, want)
	}
}
