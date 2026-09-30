package main

func TestSinkManySendersOneCloser(t *testing.T) {
	s := NewSink[int]()
	if s.Out() == nil || s.Out() != s.Out() {
		t.Fatal("Out() должен возвращать один и тот же не-nil канал")
	}
	var received atomic.Int64
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for range s.Out() {
			received.Add(1)
		}
	}()
	var accepted atomic.Int64
	panics := make(chan any, 50)
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panics <- r
				}
			}()
			for range 20000 { // предел — чтобы неверное решение не зависло навсегда
				err := s.Send(context.Background(), 1)
				if errors.Is(err, ErrClosed) {
					return
				}
				if err != nil {
					t.Errorf("Send = %v", err)
					return
				}
				accepted.Add(1)
			}
		}()
	}
	for i := 0; accepted.Load() < 200 && i < 1_000_000; i++ { // не time.Sleep: пока отправители крутятся, фиктивные часы не идут
		runtime.Gosched()
	}
	s.Close()
	wg.Wait()
	select {
	case r := <-panics:
		t.Fatalf("Send запаниковал при Close: %v", r)
	default:
	}
	select {
	case <-readerDone:
	case <-time.After(time.Second):
		t.Fatal("после Close канал Out так и не закрылся")
	}
	if accepted.Load() != received.Load() || accepted.Load() == 0 {
		t.Fatalf("Send вернул nil %d раз, а читатель получил %d — принятое значение потерялось", accepted.Load(), received.Load())
	}
}

func TestSinkCloseWakesBlockedSender(t *testing.T) {
	s := NewSink[string]()
	res := make(chan error, 1)
	go func() { res <- s.Send(context.Background(), "x") }() // читателя нет
	time.Sleep(10 * time.Millisecond)
	closed := make(chan struct{})
	go func() { s.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close завис, пока Send ждёт читателя")
	}
	select {
	case err := <-res:
		if !errors.Is(err, ErrClosed) {
			t.Fatalf("ждущий Send после Close = %v, ожидали ErrClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ждущий Send не проснулся после Close")
	}
}

func TestSinkAfterClose(t *testing.T) {
	s := NewSink[int]()
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() { defer wg.Done(); s.Close() }()
	}
	wg.Wait()
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Send после Close запаниковал: %v", r)
			}
		}()
		for range 50 {
			if err := s.Send(context.Background(), 1); !errors.Is(err, ErrClosed) {
				t.Fatalf("Send после Close = %v, ожидали ErrClosed", err)
			}
		}
	}()
	if _, ok := <-s.Out(); ok {
		t.Fatal("после Close из Out пришло значение")
	}
}

func TestSinkSendCancel(t *testing.T) {
	s := NewSink[int]()
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := s.Send(ctx, 1); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send без читателя по таймауту = %v, ожидали DeadlineExceeded", err)
	}
}
