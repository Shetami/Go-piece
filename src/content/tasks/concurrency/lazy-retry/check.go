package main

func TestLazyRetriesAfterError(t *testing.T) {
	calls := 0
	l := NewLazy(func() (string, error) {
		calls++
		if calls < 3 {
			return "", fmt.Errorf("попытка %d не удалась", calls)
		}
		return "conn", nil
	})
	for i := 1; i <= 2; i++ {
		if _, err := l.Get(); err == nil {
			t.Fatalf("вызов %d: init вернул ошибку, а Get — nil", i)
		}
	}
	for range 3 {
		v, err := l.Get()
		if err != nil || v != "conn" {
			t.Fatalf("Get = %q, %v; ожидали \"conn\", nil — ошибки не должны запоминаться", v, err)
		}
	}
	if calls != 3 {
		t.Fatalf("init вызван %d раз, ожидали 3: после успеха значение запоминается", calls)
	}
}

func TestLazyConcurrentSingleInit(t *testing.T) {
	var calls, active, peak atomic.Int32
	l := NewLazy(func() (int, error) {
		calls.Add(1)
		n := active.Add(1)
		if n > peak.Load() {
			peak.Store(n)
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		return 42, nil
	})
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := l.Get(); v != 42 || err != nil {
				t.Errorf("Get = %d, %v; ожидали 42, nil", v, err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 || peak.Load() != 1 {
		t.Fatalf("init вызван %d раз (одновременно до %d); ожидали ровно один вызов", calls.Load(), peak.Load())
	}
}

func TestLazyFailingInitNotParallel(t *testing.T) {
	var active, peak atomic.Int32
	l := NewLazy(func() (int, error) {
		n := active.Add(1)
		if n > peak.Load() {
			peak.Store(n)
		}
		time.Sleep(2 * time.Millisecond)
		active.Add(-1)
		return 0, errors.New("нет")
	})
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() { defer wg.Done(); l.Get() }()
	}
	wg.Wait()
	if peak.Load() != 1 {
		t.Fatalf("одновременно работало %d вызовов init, ожидали не больше одного", peak.Load())
	}
}

func TestLazySurvivesPanic(t *testing.T) {
	first := true
	l := NewLazy(func() (int, error) {
		if first {
			first = false
			panic("сбой инициализации")
		}
		return 7, nil
	})
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("паника init должна дойти до вызывающего Get")
			}
		}()
		l.Get()
	}()
	done := make(chan int)
	go func() { v, _ := l.Get(); done <- v }()
	select {
	case v := <-done:
		if v != 7 {
			t.Fatalf("после паники Get = %d, ожидали 7", v)
		}
	case <-time.After(time.Second):
		t.Fatal("после паники в init следующий Get завис — мьютекс остался занятым")
	}
}
