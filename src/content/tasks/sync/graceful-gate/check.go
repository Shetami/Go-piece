package main

func TestGateWaitsInflight(t *testing.T) {
	var closedAt atomic.Bool
	g := NewGate(func() error { closedAt.Store(true); return nil })
	started := make(chan struct{})
	release := make(chan struct{})
	res := make(chan error, 1)
	go func() {
		res <- g.Do(func() error {
			close(started)
			<-release
			if closedAt.Load() {
				return errors.New("ресурс закрыт посреди операции")
			}
			return nil
		})
	}()
	<-started
	closeDone := make(chan error, 1)
	go func() { closeDone <- g.Close() }()
	select {
	case <-closeDone:
		t.Fatal("Close вернулся, пока операция ещё выполнялась")
	case <-time.After(30 * time.Millisecond):
	}
	if err := g.Do(func() error { t.Error("fn вызвана после начала Close"); return nil }); !errors.Is(err, ErrClosed) {
		t.Fatalf("Do во время Close = %v, ожидали ErrClosed", err)
	}
	close(release)
	if err := <-res; err != nil {
		t.Fatalf("начатая операция: %v", err)
	}
	select {
	case <-closeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Close не вернулся после завершения операции")
	}
}

func TestGateCloseOnce(t *testing.T) {
	boom := errors.New("flush не удался")
	var calls atomic.Int32
	g := NewGate(func() error {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return boom
	})
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.Close(); !errors.Is(err, boom) {
				t.Errorf("Close = %v, ожидали ошибку onClose у каждого вызывающего", err)
			}
		}()
	}
	wg.Wait()
	if err := g.Close(); !errors.Is(err, boom) {
		t.Fatalf("повторный Close = %v, ожидали ту же ошибку", err)
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("onClose вызван %d раз, ожидали 1", n)
	}
}

func TestGateNoOpAfterClose(t *testing.T) {
	for range 30 {
		var running, violations atomic.Int32
		var closed atomic.Bool
		g := NewGate(func() error {
			if running.Load() != 0 {
				violations.Add(1)
			}
			closed.Store(true)
			return nil
		})
		var wg sync.WaitGroup
		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range 50 {
					g.Do(func() error {
						running.Add(1)
						if closed.Load() {
							violations.Add(1)
						}
						runtime.Gosched()
						running.Add(-1)
						return nil
					})
				}
			}()
		}
		runtime.Gosched()
		g.Close()
		wg.Wait()
		if v := violations.Load(); v != 0 {
			t.Fatalf("%d раз операция выполнялась одновременно с onClose или после него", v)
		}
	}
}
