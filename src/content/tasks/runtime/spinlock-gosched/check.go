package main

var _ sync.Locker = (*SpinLock)(nil)

func TestSpinLockTry(t *testing.T) {
	var l SpinLock
	if !l.TryLock() {
		t.Fatalf("TryLock открытого замка вернул false")
	}
	if l.TryLock() {
		t.Fatalf("TryLock запертого замка вернул true")
	}
	l.Unlock()
	if !l.TryLock() {
		t.Fatalf("после Unlock TryLock вернул false")
	}
	l.Unlock()
	defer func() {
		if recover() == nil {
			t.Fatalf("Unlock незапертого замка не запаниковал")
		}
	}()
	l.Unlock()
}

func TestSpinLockExclusion(t *testing.T) {
	var l SpinLock
	var inside atomic.Int32
	total := 0
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				l.Lock()
				if inside.Add(1) != 1 {
					t.Errorf("в критической секции больше одной горутины")
				}
				total++
				inside.Add(-1)
				l.Unlock()
			}
		}()
	}
	wg.Wait()
	if total != 4000 {
		t.Fatalf("счётчик = %d, ожидали 4000 — замок пропускает двоих", total)
	}
}

func TestSpinLockYields(t *testing.T) {
	prev := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(prev)
	var l SpinLock
	done := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		for g := 0; g < 4; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					l.Lock()
					runtime.Gosched() // владелец уступает процессор внутри секции
					l.Unlock()
				}
			}()
		}
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("GOMAXPROCS=1: 400 захватов не уложились в 2 с — Lock крутится, не уступая процессор владельцу")
	}
}
