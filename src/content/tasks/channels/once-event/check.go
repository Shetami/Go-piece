package main

func TestEventWakesAll(t *testing.T) {
	e := NewEvent()
	errs := make(chan error, 5)
	for range 5 {
		go func() { errs <- e.Wait(context.Background()) }()
	}
	time.Sleep(20 * time.Millisecond)
	select {
	case err := <-errs:
		t.Fatalf("Wait вернулся (%v) до Fire", err)
	default:
	}
	e.Fire()
	for i := range 5 {
		select {
		case err := <-errs:
			if err != nil {
				t.Fatalf("Wait после Fire вернул %v, ожидали nil", err)
			}
		case <-time.After(time.Second):
			t.Fatalf("после Fire проснулись только %d из 5 ожидающих", i)
		}
	}
	select {
	case <-e.Done():
	case <-time.After(time.Second):
		t.Fatal("после Fire канал Done не закрыт (или Done вернул nil)")
	}
	if e.Done() != e.Done() {
		t.Fatal("Done должен каждый раз возвращать один и тот же канал")
	}
}

func TestEventFireConcurrent(t *testing.T) {
	e := NewEvent()
	panics := make(chan any, 10)
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics <- r
				}
			}()
			e.Fire()
		}()
	}
	wg.Wait()
	select {
	case r := <-panics:
		t.Fatalf("повторный Fire запаниковал: %v", r)
	default:
	}
}

func TestEventWaitCancel(t *testing.T) {
	e := NewEvent()
	ctx, cancel := context.WithCancel(context.Background())
	res := make(chan error, 1)
	go func() { res <- e.Wait(ctx) }()
	cancel()
	select {
	case err := <-res:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Wait при отмене вернул %v, ожидали context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Wait не вернулся после отмены ctx")
	}
}

func TestEventFiredBeatsCancel(t *testing.T) {
	e := NewEvent()
	e.Fire()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := range 100 {
		if err := e.Wait(ctx); err != nil {
			t.Fatalf("попытка %d: событие уже наступило, а Wait вернул %v", i, err)
		}
	}
}
