package main

func TestFirstOfFastestWins(t *testing.T) {
	canceled := make(chan struct{})
	slow := func(ctx context.Context) (string, error) {
		<-ctx.Done()
		close(canceled)
		return "", ctx.Err()
	}
	fast := func(ctx context.Context) (string, error) { return "быстрый", nil }

	got, err := FirstOf(context.Background(), slow, fast)
	if err != nil || got != "быстрый" {
		t.Fatalf("FirstOf = %q, %v; ожидали быстрый ответ", got, err)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("медленная функция так и не получила отмену контекста")
	}
}

func TestFirstOfSkipsFailures(t *testing.T) {
	fail := func(ctx context.Context) (string, error) { return "", errors.New("упал") }
	ok := func(ctx context.Context) (string, error) {
		time.Sleep(20 * time.Millisecond)
		return "выжил", nil
	}
	got, err := FirstOf(context.Background(), fail, ok, fail)
	if err != nil || got != "выжил" {
		t.Fatalf("FirstOf = %q, %v; ошибки должны пропускаться, пока есть надежда", got, err)
	}
}

func TestFirstOfAllFail(t *testing.T) {
	e1, e2 := errors.New("диск"), errors.New("сеть")
	_, err := FirstOf(context.Background(),
		func(context.Context) (string, error) { return "", e1 },
		func(context.Context) (string, error) { return "", e2 },
	)
	if !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("когда упали все, ожидали обе ошибки, получили %v", err)
	}
}
