package main

var (
	chkErrBiz      = errors.New("товар закончился")
	chkErrShutdown = errors.New("сервис останавливается")
)

func chkWait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		return errors.New("контекст так и не завершился")
	}
}

func TestCallOK(t *testing.T) {
	var got context.Context
	err := Call(context.Background(), "склад", time.Hour, func(ctx context.Context) error { got = ctx; return nil })
	if err != nil {
		t.Fatalf("успех: ожидали nil, получили %v", err)
	}
	if got == nil || got.Err() == nil {
		t.Fatal("после возврата Call контекст f должен быть отменён (забыли cancel?)")
	}
	if _, ok := got.Deadline(); !ok {
		t.Fatal("у контекста f должен быть дедлайн")
	}
}

func TestCallBusinessError(t *testing.T) {
	err := Call(context.Background(), "склад", time.Hour, func(context.Context) error { return chkErrBiz })
	if !errors.Is(err, chkErrBiz) || err.Error() != "склад: товар закончился" {
		t.Fatalf("бизнес-ошибка: получили %v, ожидали «склад: товар закончился»", err)
	}
	var te *TimeoutError
	if errors.As(err, &te) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("обычная ошибка не должна выглядеть как таймаут")
	}
}

func TestCallOwnTimeout(t *testing.T) {
	err := Call(context.Background(), "склад", 20*time.Millisecond, chkWait)
	var te *TimeoutError
	if !errors.As(err, &te) || te.Op != "склад" || te.After != 20*time.Millisecond {
		t.Fatalf("наш таймаут: ожидали *TimeoutError{склад, 20ms}, получили %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("*TimeoutError должен отвечать на errors.Is(err, context.DeadlineExceeded)")
	}

	// f игнорирует контекст и возвращает свою ошибку уже после нашего таймаута.
	err = Call(context.Background(), "склад", 10*time.Millisecond, func(context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return chkErrBiz
	})
	if !errors.As(err, &te) || !errors.Is(err, chkErrBiz) {
		t.Fatalf("таймаут истёк, пока f работала: ожидали *TimeoutError с ошибкой f внутри, получили %v", err)
	}
}

func TestCallParentDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := Call(ctx, "склад", time.Hour, chkWait)
	var te *TimeoutError
	if errors.As(err, &te) {
		t.Fatalf("истёк дедлайн вызывающего, а не наш часовой таймаут — *TimeoutError тут неуместен: %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "склад") {
		t.Fatalf("ожидали DeadlineExceeded с op в тексте, получили %v", err)
	}
}

func TestCallParentCause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel(chkErrShutdown) }()
	err := Call(ctx, "склад", time.Hour, chkWait)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, chkErrShutdown) {
		t.Fatalf("отмена родителя с причиной: ожидали и context.Canceled, и причину, получили %v", err)
	}
}
