package main

func chkCloseAll(t *testing.T, c *Conn, n int) []error {
	t.Helper()
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errs[i] = fmt.Errorf("Close запаниковал: %v", r)
				}
			}()
			errs[i] = c.Close()
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("одновременные Close зависли")
	}
	return errs
}

func TestConnCloseOnceSameError(t *testing.T) {
	var calls atomic.Int32
	flushErr := errors.New("буфер не сброшен")
	c := NewConn(func() error { calls.Add(1); time.Sleep(10 * time.Millisecond); return flushErr })
	errs := chkCloseAll(t, c, 20)
	for i, err := range errs {
		if err != flushErr {
			t.Fatalf("Close #%d = %v, ожидали ошибку closeFn у всех — и у тех, кто ждал", i, err)
		}
	}
	if err := c.Close(); err != flushErr {
		t.Fatalf("повторный Close = %v, ожидали ту же ошибку", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("closeFn вызвана %d раз, ожидали 1", calls.Load())
	}
}

func TestConnCloseSuccess(t *testing.T) {
	c := NewConn(func() error { return nil })
	select {
	case <-c.Done():
		t.Fatal("Done закрыт до Close")
	default:
	}
	if err := c.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	select {
	case <-c.Done():
	case <-time.After(time.Second):
		t.Fatal("после Close канал Done должен быть закрыт")
	}
}

func TestConnClosePanics(t *testing.T) {
	var calls atomic.Int32
	c := NewConn(func() error { calls.Add(1); time.Sleep(5 * time.Millisecond); panic("сокет уже освобождён") })
	errs := chkCloseAll(t, c, 10)
	first := errs[0]
	if first == nil || !strings.Contains(first.Error(), "сокет уже освобождён") || strings.Contains(first.Error(), "Close запаниковал") {
		t.Fatalf("Close = %v, ожидали ошибку с текстом паники, а не панику", first)
	}
	for i, err := range append(errs, c.Close()) {
		if err == nil || err.Error() != first.Error() {
			t.Fatalf("Close #%d = %v, ожидали ту же ошибку %q — паника closeFn не должна превращаться в nil", i, err, first)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("closeFn вызвана %d раз, ожидали 1", calls.Load())
	}
	select {
	case <-c.Done():
	case <-time.After(time.Second):
		t.Fatal("после паники closeFn канал Done тоже должен закрыться")
	}
}
