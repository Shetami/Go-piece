package main

func TestSleepFull(t *testing.T) {
	start := time.Now()
	if err := Sleep(context.Background(), 30*time.Millisecond); err != nil {
		t.Fatalf("Sleep без отмены: %v", err)
	}
	if time.Since(start) < 30*time.Millisecond {
		t.Fatal("Sleep вернулся раньше срока")
	}
}

func TestSleepCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := Sleep(ctx, 5*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ожидали DeadlineExceeded, получили %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("Sleep не прервался по отмене контекста")
	}
}

func TestSleepAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("уже отменённый контекст: ожидали Canceled, получили %v", err)
	}
}
