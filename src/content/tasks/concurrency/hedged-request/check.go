package main

func TestHedgedFastPrimaryNoHedge(t *testing.T) {
	var calls atomic.Int32
	v, err := Hedged(context.Background(), 3, 200*time.Millisecond, func(ctx context.Context, i int) (string, error) {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		return fmt.Sprint("replica-", i), nil
	})
	if err != nil || v != "replica-0" {
		t.Fatalf("Hedged = %q, %v; ожидали \"replica-0\", nil", v, err)
	}
	time.Sleep(300 * time.Millisecond)
	if n := calls.Load(); n != 1 {
		t.Fatalf("быстрый первый ответ, а попыток %d — страховка не нужна", n)
	}
}

func TestHedgedSlowPrimary(t *testing.T) {
	primaryCanceled := make(chan struct{})
	v, err := Hedged(context.Background(), 3, 50*time.Millisecond, func(ctx context.Context, i int) (int, error) {
		if i == 0 {
			<-ctx.Done() // зависшая реплика
			close(primaryCanceled)
			return 0, ctx.Err()
		}
		return 100 + i, nil
	})
	if err != nil || v != 101 {
		t.Fatalf("Hedged = %d, %v; ожидали 101 от второй попытки", v, err)
	}
	select {
	case <-primaryCanceled:
	case <-time.After(time.Second):
		t.Fatal("проигравшая попытка не получила отмену контекста")
	}
}

func TestHedgedErrorStartsNextImmediately(t *testing.T) {
	start := time.Now()
	v, err := Hedged(context.Background(), 3, 2*time.Second, func(ctx context.Context, i int) (int, error) {
		if i == 0 {
			return 0, errors.New("503")
		}
		return 7, nil
	})
	if err != nil || v != 7 {
		t.Fatalf("Hedged = %d, %v; ожидали 7, nil", v, err)
	}
	if el := time.Since(start); el > time.Second {
		t.Fatalf("ушло %v — после ошибки следующая попытка ждала delay", el)
	}
}

func TestHedgedAllFailOrdered(t *testing.T) {
	e := []error{errors.New("e0"), errors.New("e1"), errors.New("e2")}
	_, err := Hedged(context.Background(), 3, 20*time.Millisecond, func(ctx context.Context, i int) (int, error) {
		time.Sleep(time.Duration(3-i) * 60 * time.Millisecond) // первая падает последней
		return 0, e[i]
	})
	for _, x := range e {
		if !errors.Is(err, x) {
			t.Fatalf("в ошибке %v нет %v — нужны все ошибки попыток", err, x)
		}
	}
	if err.Error() != "e0\ne1\ne2" {
		t.Fatalf("ошибка %q, ожидали \"e0\\ne1\\ne2\" — порядок по номерам попыток", err.Error())
	}
}

func TestHedgedParentCancelNoLeak(t *testing.T) {
	base := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := Hedged(ctx, 3, 10*time.Millisecond, func(ctx context.Context, i int) (int, error) {
			<-release // не слушает ctx: вернётся, когда отпустим
			return i, nil
		})
		done <- err
	}()
	time.Sleep(80 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("после отмены Hedged вернул %v, ожидали context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("после отмены ctx Hedged не вернулся")
	}
	close(release) // попытки заканчиваются, когда их результат уже никому не нужен
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine() - base; n > 0 {
		t.Fatalf("осталось %d горутин, застрявших на отправке результата", n)
	}
}
