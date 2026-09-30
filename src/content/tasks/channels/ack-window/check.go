package main

func chkNoBlock(t *testing.T, what string, f func() error) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- f() }()
	select {
	case err := <-ch:
		return err
	case <-time.After(time.Second):
		t.Fatalf("%s заблокировался", what)
	}
	return nil
}

func chkTimeout(t *testing.T, ms int) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ms)*time.Millisecond)
	t.Cleanup(cancel)
	return ctx
}

func TestWindowBlocksWhenFull(t *testing.T) {
	w := NewWindow(2)
	bg := context.Background()
	if err := w.Acquire(bg, "a"); err != nil {
		t.Fatalf("Acquire(a) = %v", err)
	}
	if err := w.Acquire(bg, "b"); err != nil {
		t.Fatalf("Acquire(b) = %v", err)
	}
	if err := w.Acquire(chkTimeout(t, 20), "c"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("окно полно, а Acquire(c) = %v; ожидали DeadlineExceeded", err)
	}
	done := make(chan error, 1)
	go func() { done <- w.Acquire(bg, "c") }()
	time.Sleep(10 * time.Millisecond)
	if err := chkNoBlock(t, "Ack(a)", func() error { return w.Ack("a") }); err != nil {
		t.Fatalf("Ack(a) = %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Acquire(c) после Ack = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Ack освободил место, а Acquire(c) так и ждёт")
	}
	if got := w.InFlight(); !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatalf("InFlight = %v, ожидали [b c]", got)
	}
}

func TestWindowDoubleAck(t *testing.T) {
	w := NewWindow(1)
	w.Acquire(context.Background(), "a")
	if err := w.Ack("a"); err != nil {
		t.Fatalf("Ack(a) = %v", err)
	}
	err := chkNoBlock(t, "повторный Ack(a)", func() error { return w.Ack("a") })
	if !errors.Is(err, ErrUnknownID) {
		t.Fatalf("повторный Ack(a) = %v, ожидали ErrUnknownID", err)
	}
	if err := chkNoBlock(t, "Ack(zzz) на пустом окне", func() error { return w.Ack("zzz") }); !errors.Is(err, ErrUnknownID) {
		t.Fatalf("Ack неизвестного id = %v, ожидали ErrUnknownID", err)
	}
	if err := w.Acquire(chkTimeout(t, 500), "b"); err != nil {
		t.Fatalf("Acquire(b) = %v", err)
	}
	if err := w.Acquire(chkTimeout(t, 20), "c"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("size = 1 и b в окне, а Acquire(c) = %v: лишний Ack расширил окно", err)
	}
}

func TestWindowDuplicate(t *testing.T) {
	w := NewWindow(2)
	w.Acquire(context.Background(), "a")
	if err := w.Acquire(chkTimeout(t, 500), "a"); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("повторный Acquire(a) = %v, ожидали ErrDuplicate", err)
	}
	if err := w.Acquire(chkTimeout(t, 500), "b"); err != nil {
		t.Fatalf("после отказа по дублю место должно остаться: Acquire(b) = %v", err)
	}
	if got := w.InFlight(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("InFlight = %v, ожидали [a b]", got)
	}
}

func TestWindowConcurrentLimit(t *testing.T) {
	w := NewWindow(3)
	var cur, peak atomic.Int64
	var wg sync.WaitGroup
	for g := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 20 {
				id := fmt.Sprintf("%d-%d", g, i)
				if err := w.Acquire(context.Background(), id); err != nil {
					t.Errorf("Acquire(%s) = %v", id, err)
					return
				}
				n := cur.Add(1)
				for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
				}
				time.Sleep(50 * time.Microsecond)
				cur.Add(-1)
				if err := w.Ack(id); err != nil {
					t.Errorf("Ack(%s) = %v", id, err)
					return
				}
			}
		}()
	}
	wg.Wait()
	if p := peak.Load(); p > 3 {
		t.Fatalf("в полёте одновременно было %d сообщений при size = 3", p)
	}
	if n := len(w.InFlight()); n != 0 {
		t.Fatalf("после всех Ack в окне осталось %d id", n)
	}
}
