package main

func chkRecover(f func()) (p any) {
	defer func() { p = recover() }()
	f()
	return nil
}

func TestOnceValuesConcurrent(t *testing.T) {
	var calls atomic.Int32
	get := OnceValues(func() (int, error) {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})
	var wg sync.WaitGroup
	var bad atomic.Int32
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := get(); v != 42 || err != nil {
				bad.Add(1)
			}
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("f вызвана %d раз из 50 горутин, ожидали ровно 1", n)
	}
	if bad.Load() != 0 {
		t.Fatalf("%d горутин получили не (42, nil) — кто-то не дождался первого вызова", bad.Load())
	}
}

func TestOnceValuesCachesError(t *testing.T) {
	boom := errors.New("нет конфига")
	calls := 0
	get := OnceValues(func() (string, error) {
		calls++
		return "", boom
	})
	for range 3 {
		if _, err := get(); !errors.Is(err, boom) {
			t.Fatalf("ожидали запомненную ошибку %v, получили %v", boom, err)
		}
	}
	if calls != 1 {
		t.Fatalf("f с ошибкой вызвана %d раз, ожидали 1: ошибка тоже запоминается", calls)
	}
}

func TestOnceValuesPanicEveryTime(t *testing.T) {
	calls := 0
	get := OnceValues(func() (int, error) {
		calls++
		panic("сломалось")
	})
	for i := range 3 {
		p := chkRecover(func() { get() })
		if p != "сломалось" {
			t.Fatalf("вызов %d: паника %v, ожидали \"сломалось\" на каждом вызове", i+1, p)
		}
	}
	if calls != 1 {
		t.Fatalf("паникующая f вызвана %d раз, ожидали 1", calls)
	}
}

func TestOnceValuesIndependent(t *testing.T) {
	a := OnceValues(func() (int, error) { return 1, nil })
	b := OnceValues(func() (int, error) { return 2, nil })
	if x, _ := a(); x != 1 {
		t.Fatalf("a() = %d, ожидали 1", x)
	}
	if x, _ := b(); x != 2 {
		t.Fatalf("b() = %d, ожидали 2: у каждой обёртки своё состояние", x)
	}
}
