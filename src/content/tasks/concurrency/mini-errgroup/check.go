package main

func TestGroupWaitsAll(t *testing.T) {
	g, _ := WithContext(context.Background())
	var done atomic.Int64
	for range 5 {
		g.Go(func() error {
			time.Sleep(10 * time.Millisecond)
			done.Add(1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if done.Load() != 5 {
		t.Fatalf("Wait вернулся, когда закончилось %d задач из 5", done.Load())
	}
}

func TestGroupFirstErrorCancels(t *testing.T) {
	g, ctx := WithContext(context.Background())
	boom := errors.New("первая")
	g.Go(func() error { return boom })
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return errors.New("отменили")
		case <-time.After(2 * time.Second):
			return errors.New("отмена так и не пришла")
		}
	})
	if err := g.Wait(); !errors.Is(err, boom) {
		t.Fatalf("Wait должен вернуть первую ошибку, получили %v", err)
	}
	if ctx.Err() == nil {
		t.Fatal("после ошибки контекст группы должен быть отменён")
	}
}

func TestGroupNoErrorContextCanceledAfterWait(t *testing.T) {
	g, ctx := WithContext(context.Background())
	g.Go(func() error { return nil })
	g.Wait()
	if ctx.Err() == nil {
		t.Fatal("после Wait контекст группы отменяется — как у errgroup")
	}
}
