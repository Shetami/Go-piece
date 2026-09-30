package main

var (
	chkErrDB       = errors.New("база недоступна")
	chkErrRollback = errors.New("откат не удался")
	chkErrNoCancel = errors.New("не дождались отмены контекста")
)

// chkWaitCtx ждёт отмены, но не дольше двух секунд — чтобы тест не зависал.
func chkWaitCtx(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(2 * time.Second):
		return false
	}
}

func TestGroupNoErrors(t *testing.T) {
	g, ctx := WithContext(context.Background())
	var done atomic.Int64
	for range 5 {
		g.Go(func() error { time.Sleep(10 * time.Millisecond); done.Add(1); return nil })
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("ошибок не было, Wait = %v", err)
	}
	if done.Load() != 5 {
		t.Fatalf("Wait вернулся, когда закончились %d задач из 5", done.Load())
	}
	if ctx.Err() == nil {
		t.Fatal("после Wait контекст группы должен быть отменён")
	}
}

func TestGroupRootCause(t *testing.T) {
	g, ctx := WithContext(context.Background())
	var cause atomic.Value
	g.Go(func() error { return chkErrDB })
	for range 3 {
		g.Go(func() error {
			if !chkWaitCtx(ctx) {
				return chkErrNoCancel
			}
			cause.Store(context.Cause(ctx))
			return ctx.Err()
		})
	}
	g.Go(func() error {
		if !chkWaitCtx(ctx) {
			return chkErrNoCancel
		}
		return fmt.Errorf("сохранение прервано: %w", context.Cause(ctx))
	})
	err := g.Wait()
	if errors.Is(err, chkErrNoCancel) {
		t.Fatal("после первой ошибки контекст группы не отменился")
	}
	if c, _ := cause.Load().(error); c != chkErrDB {
		t.Fatalf("context.Cause(ctx) = %v, ожидали первую ошибку %q", c, chkErrDB)
	}
	if err == nil || err.Error() != chkErrDB.Error() {
		t.Fatalf("Wait = %q, ожидали только первопричину %q — эхо отмены надо отбросить", err, chkErrDB)
	}
	if errors.Is(err, context.Canceled) {
		t.Fatal("в итоговой ошибке не должно быть context.Canceled от отменённых задач")
	}
}

func TestGroupKeepsRealErrors(t *testing.T) {
	g, ctx := WithContext(context.Background())
	g.Go(func() error { return chkErrDB })
	g.Go(func() error {
		if !chkWaitCtx(ctx) {
			return chkErrNoCancel
		}
		return chkErrRollback // настоящий новый сбой, случившийся при отмене
	})
	err := g.Wait()
	if !errors.Is(err, chkErrDB) || !errors.Is(err, chkErrRollback) {
		t.Fatalf("Wait = %q, ожидали и первопричину, и ошибку отката", err)
	}
	if first, _, _ := strings.Cut(err.Error(), "\n"); first != chkErrDB.Error() {
		t.Fatalf("первой строкой должна идти первопричина, а там %q", first)
	}
}

func TestGroupConcurrentErrors(t *testing.T) {
	g, ctx := WithContext(context.Background())
	errs := make([]error, 50)
	for i := range errs {
		errs[i] = fmt.Errorf("сбой %d", i)
		g.Go(func() error { return errs[i] })
	}
	err := g.Wait()
	for _, e := range errs {
		if !errors.Is(err, e) {
			t.Fatalf("потеряна ошибка %q", e)
		}
	}
	if first, _, _ := strings.Cut(err.Error(), "\n"); first != context.Cause(ctx).Error() {
		t.Fatalf("первая строка %q, а причина отмены контекста %q — должны совпадать", first, context.Cause(ctx))
	}
}
