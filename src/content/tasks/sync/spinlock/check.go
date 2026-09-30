package main

func TestSpinLockExclusive(t *testing.T) {
	var l SpinLock
	var inside atomic.Int32
	counter := 0
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 2000 {
				l.Lock()
				if n := inside.Add(1); n != 1 {
					t.Errorf("внутри критической секции %d горутин одновременно", n)
				}
				counter++
				inside.Add(-1)
				l.Unlock()
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("8 горутин не смогли по очереди пройти через замок — зависание")
	}
	if counter != 16000 {
		t.Fatalf("counter = %d, ожидали 16000", counter)
	}
}

func TestSpinLockTryLock(t *testing.T) {
	var l SpinLock
	if !l.TryLock() {
		t.Fatal("TryLock на открытом замке вернул false")
	}
	if l.TryLock() {
		t.Fatal("TryLock на закрытом замке вернул true")
	}
	l.Unlock()
	if !l.TryLock() {
		t.Fatal("после Unlock TryLock вернул false")
	}
	l.Unlock()
}

func TestSpinLockUnlockUnlocked(t *testing.T) {
	var l SpinLock
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(fmt.Sprint(r), "unlock of unlocked SpinLock") {
			t.Fatalf("Unlock открытого замка: recover() = %v, ожидали панику \"unlock of unlocked SpinLock\"", r)
		}
	}()
	l.Unlock()
}

func TestSpinLockWithCond(t *testing.T) {
	var l SpinLock
	var locker sync.Locker = &l
	cond := sync.NewCond(locker)
	ready := false
	got := make(chan bool)
	go func() {
		l.Lock()
		for !ready {
			cond.Wait()
		}
		l.Unlock()
		got <- true
	}()
	l.Lock()
	ready = true
	cond.Broadcast()
	l.Unlock()
	select {
	case <-got:
	case <-time.After(5 * time.Second):
		t.Fatal("sync.Cond поверх SpinLock не разбудил ждущего")
	}
	if !l.TryLock() {
		t.Fatal("после работы с Cond замок остался закрытым")
	}
}
