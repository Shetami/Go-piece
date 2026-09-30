package main

func chkGet[T any](t *testing.T, l *Lazy[T]) (v T, err error, rec any) {
	t.Helper()
	type out struct {
		v   T
		err error
		rec any
	}
	ch := make(chan out, 1)
	go func() {
		var o out
		defer func() { o.rec = recover(); ch <- o }()
		o.v, o.err = l.Get()
	}()
	select {
	case o := <-ch:
		return o.v, o.err, o.rec
	case <-time.After(2 * time.Second):
		t.Fatal("Get завис — мьютекс не отпущен после паники?")
		return
	}
}

func TestLazyCachesSuccess(t *testing.T) {
	calls := 0
	l := NewLazy(func() (string, error) { calls++; return "dsn", nil })
	for range 3 {
		if v, err, _ := chkGet(t, l); v != "dsn" || err != nil {
			t.Fatalf("Get() = %q, %v; ожидали \"dsn\", nil", v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("init вызвана %d раз, ожидали 1", calls)
	}
}

func TestLazyRetriesAfterError(t *testing.T) {
	calls := 0
	down := errors.New("конфиг-сервер недоступен")
	l := NewLazy(func() (int, error) {
		calls++
		if calls == 1 {
			return 0, down
		}
		return 7, nil
	})
	if _, err, _ := chkGet(t, l); !errors.Is(err, down) {
		t.Fatalf("первый Get: err = %v, ожидали ошибку init", err)
	}
	if v, err, _ := chkGet(t, l); v != 7 || err != nil {
		t.Fatalf("после ошибки Get() = %d, %v; ожидали повторный init и 7, nil", v, err)
	}
}

func TestLazyRetriesAfterPanic(t *testing.T) {
	calls := 0
	l := NewLazy(func() (int, error) {
		calls++
		if calls <= 2 {
			panic("init упал")
		}
		return 42, nil
	})
	for i := 1; i <= 2; i++ {
		if _, _, rec := chkGet(t, l); rec != "init упал" {
			t.Fatalf("Get #%d: паника init должна долететь до вызывающего, recover() = %v", i, rec)
		}
	}
	v, err, rec := chkGet(t, l)
	if v != 42 || err != nil || rec != nil {
		t.Fatalf("после паник Get() = %d, %v, паника %v; ожидали 42, nil — Lazy должен остаться рабочим", v, err, rec)
	}
	if calls != 3 {
		t.Fatalf("init вызвана %d раз, ожидали 3", calls)
	}
}

func TestLazyConcurrentSingleInit(t *testing.T) {
	var calls, running, overlap atomic.Int32
	l := NewLazy(func() (int, error) {
		if running.Add(1) > 1 {
			overlap.Store(1)
		}
		defer running.Add(-1)
		n := calls.Add(1)
		time.Sleep(5 * time.Millisecond)
		if n == 1 {
			panic("первый упал")
		}
		return 1, nil
	})
	var wg sync.WaitGroup
	var ok atomic.Int32
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { recover() }()
			if v, err := l.Get(); v == 1 && err == nil {
				ok.Add(1)
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("горутины зависли в Get")
	}
	if overlap.Load() != 0 {
		t.Fatal("init выполнялась в нескольких горутинах одновременно")
	}
	if c := calls.Load(); c != 2 {
		t.Fatalf("init вызвана %d раз, ожидали 2: одна паника и один успех", c)
	}
	if ok.Load() != 49 {
		t.Fatalf("значение получили %d горутин из 49 не запаниковавших", ok.Load())
	}
}
