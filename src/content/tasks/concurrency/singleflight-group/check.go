package main

func chkWaitCount(c *atomic.Int32, n int32) {
	for i := 0; i < 200 && c.Load() < n; i++ {
		time.Sleep(5 * time.Millisecond)
	}
}

func TestGroupDedup(t *testing.T) {
	var g Group[string, string]
	var calls atomic.Int32
	release := make(chan struct{})
	fn := func() (string, error) {
		calls.Add(1)
		<-release
		return "v", nil
	}
	type res struct {
		v      string
		err    error
		shared bool
	}
	out := make(chan res, 10)
	for range 10 {
		go func() {
			v, err, sh := g.Do("k", fn)
			out <- res{v, err, sh}
		}()
	}
	chkWaitCount(&calls, 1)
	time.Sleep(50 * time.Millisecond) // остальные успевают присоединиться
	close(release)
	for range 10 {
		select {
		case r := <-out:
			if r.v != "v" || r.err != nil || !r.shared {
				t.Fatalf("Do = %q, %v, shared=%v; ожидали \"v\", nil, true", r.v, r.err, r.shared)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("не все вызовы Do вернулись — ждущие не проснулись")
		}
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("fn вызвана %d раз, ожидали 1", n)
	}
}

func TestGroupForgetsKeyAfterCall(t *testing.T) {
	var g Group[string, int]
	calls := 0
	fn := func() (int, error) { calls++; return calls, nil }
	v1, _, sh1 := g.Do("k", fn)
	v2, _, sh2 := g.Do("k", fn)
	if v1 != 1 || v2 != 2 || sh1 || sh2 {
		t.Fatalf("последовательные Do: (%d, %v), (%d, %v); ожидали (1, false), (2, false) — результат не кэшируется", v1, sh1, v2, sh2)
	}
}

func TestGroupSharesError(t *testing.T) {
	var g Group[int, int]
	boom := errors.New("база недоступна")
	var calls atomic.Int32
	release := make(chan struct{})
	errs := make(chan error, 5)
	for range 5 {
		go func() {
			_, err, _ := g.Do(1, func() (int, error) { calls.Add(1); <-release; return 0, boom })
			errs <- err
		}()
	}
	chkWaitCount(&calls, 1)
	time.Sleep(50 * time.Millisecond)
	close(release)
	for range 5 {
		if err := <-errs; !errors.Is(err, boom) {
			t.Fatalf("ждущий получил %v, ожидали ошибку fn", err)
		}
	}
}

func TestGroupKeysIndependent(t *testing.T) {
	var g Group[string, int]
	bDone := make(chan struct{})
	go g.Do("a", func() (int, error) { <-bDone; return 1, nil })
	time.Sleep(10 * time.Millisecond)
	finished := make(chan struct{})
	go func() {
		g.Do("b", func() (int, error) { return 2, nil })
		close(bDone)
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Do(\"b\") ждёт, пока закончится Do(\"a\") — разные ключи не должны блокировать друг друга")
	}
}

func TestGroupPanic(t *testing.T) {
	var g Group[string, int]
	var calls atomic.Int32
	release := make(chan struct{})
	leader := make(chan any, 1)
	go func() {
		defer func() { leader <- recover() }()
		g.Do("p", func() (int, error) { calls.Add(1); <-release; panic("упали") })
	}()
	chkWaitCount(&calls, 1)
	waiter := make(chan error, 1)
	go func() { _, err, _ := g.Do("p", func() (int, error) { return 5, nil }); waiter <- err }()
	time.Sleep(50 * time.Millisecond)
	close(release)
	if r := <-leader; r == nil {
		t.Fatal("паника fn должна продолжиться в горутине, которая её запускала")
	}
	select {
	case err := <-waiter:
		if !errors.Is(err, ErrPanicked) {
			t.Fatalf("ждущий получил %v, ожидали ErrPanicked", err)
		}
	case <-time.After(time.Second):
		t.Fatal("после паники fn ждущий вызов Do завис навсегда")
	}
	done := make(chan int, 1)
	go func() { v, _, _ := g.Do("p", func() (int, error) { return 9, nil }); done <- v }()
	select {
	case v := <-done:
		if v != 9 {
			t.Fatalf("после паники Do = %d, ожидали 9 — ключ должен освободиться", v)
		}
	case <-time.After(time.Second):
		t.Fatal("после паники ключ так и остался занят")
	}
}
