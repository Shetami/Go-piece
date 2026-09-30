package main

func chkBg(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestStoreConcurrentUpdates(t *testing.T) {
	s := NewStore()
	defer s.Close()
	var wg sync.WaitGroup
	for g := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				if _, err := s.Update(chkBg(t), fmt.Sprint("k", g%4), func(o int) int { return o + 1 }); err != nil {
					t.Errorf("Update = %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	snap, err := s.Snapshot(chkBg(t))
	want := map[string]int{"k0": 250, "k1": 250, "k2": 250, "k3": 250}
	if err != nil || !reflect.DeepEqual(snap, want) {
		t.Fatalf("Snapshot = %v, %v; ожидали %v", snap, err, want)
	}
	if v, _ := s.Update(chkBg(t), "k0", func(o int) int { return o * 2 }); v != 500 {
		t.Fatalf("Update вернул %d, ожидали новое значение 500", v)
	}
}

func TestStoreSnapshotIsCopy(t *testing.T) {
	s := NewStore()
	defer s.Close()
	s.Update(chkBg(t), "a", func(int) int { return 1 })
	snap, _ := s.Snapshot(chkBg(t))
	if snap == nil {
		t.Fatal("Snapshot вернул nil")
	}
	snap["a"] = 100
	snap["b"] = 7
	again, _ := s.Snapshot(chkBg(t))
	if !reflect.DeepEqual(again, map[string]int{"a": 1}) {
		t.Fatalf("после изменения снимка хранилище стало %v — Snapshot отдал внутреннюю мапу", again)
	}
}

func TestStoreCallerLeaves(t *testing.T) {
	s := NewStore()
	// без defer Close: у неверного решения владелец завис бы, и Close тоже
	ctx, cancel := context.WithCancel(context.Background())
	_, err := s.Update(ctx, "x", func(o int) int {
		cancel()                          // вызывающий уходит, пока владелец ещё считает
		time.Sleep(20 * time.Millisecond) // и успевает уйти до ответа
		return o + 1
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Update с отменой во время обработки = %v, ожидали context.Canceled", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	v, err := s.Update(ctx2, "x", func(o int) int { return o + 1 })
	if err != nil {
		t.Fatalf("после ушедшего вызывающего хранилище зависло: Update = %v", err)
	}
	if v != 2 {
		t.Fatalf("x = %d, ожидали 2 (первое изменение успело примениться)", v)
	}
}

func TestStoreClose(t *testing.T) {
	before := runtime.NumGoroutine()
	s := NewStore()
	s.Update(chkBg(t), "a", func(int) int { return 1 })
	done := make(chan struct{})
	go func() {
		s.Close()
		s.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Close (дважды) не вернулся")
	}
	if _, err := s.Update(chkBg(t), "a", func(o int) int { return o }); !errors.Is(err, ErrClosed) {
		t.Fatalf("Update после Close = %v, ожидали ErrClosed", err)
	}
	if _, err := s.Snapshot(chkBg(t)); !errors.Is(err, ErrClosed) {
		t.Fatalf("Snapshot после Close = %v, ожидали ErrClosed", err)
	}
	time.Sleep(10 * time.Millisecond)
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("после Close осталась горутина-владелец (%d лишних)", n-before)
	}
}
