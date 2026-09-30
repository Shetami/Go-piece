package main

func TestKeyedMutexExclusivePerKey(t *testing.T) {
	var km KeyedMutex
	keys := []string{"order-1", "order-2", "order-3"}
	var inside [3]atomic.Int32
	var reported atomic.Bool
	var counts [3]int
	var wg sync.WaitGroup
	for g := range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 300 {
				k := (g + i) % 3
				unlock := km.Lock(keys[k])
				if n := inside[k].Add(1); n != 1 && reported.CompareAndSwap(false, true) {
					t.Errorf("ключ %s держат %d горутин одновременно", keys[k], n)
				}
				counts[k]++
				runtime.Gosched()
				inside[k].Add(-1)
				unlock()
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("12 горутин на трёх ключах зависли")
	}
	if counts[0]+counts[1]+counts[2] != 3600 {
		t.Fatalf("всего входов %v, ожидали 3600", counts)
	}
	if n := km.Len(); n != 0 {
		t.Fatalf("после всех unlock Len() = %d, ожидали 0 — записи о ключах не удаляются", n)
	}
}

func TestKeyedMutexKeysIndependent(t *testing.T) {
	var km KeyedMutex
	unlockA := km.Lock("a")
	done := make(chan struct{})
	go func() { km.Lock("b")(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Lock(b) ждёт, пока держат a")
	}
	blocked := make(chan struct{})
	go func() { km.Lock("a")(); close(blocked) }()
	select {
	case <-blocked:
		t.Fatal("второй Lock(a) прошёл, пока a держат")
	case <-time.After(30 * time.Millisecond):
	}
	if n := km.Len(); n != 1 {
		t.Fatalf("держат a и ещё один ждёт: Len() = %d, ожидали 1", n)
	}
	unlockA()
	unlockA() // повторный unlock — ничего не делает
	select {
	case <-blocked:
	case <-time.After(time.Second):
		t.Fatal("после unlock ждущий Lock(a) не прошёл")
	}
	if n := km.Len(); n != 0 {
		t.Fatalf("Len() = %d, ожидали 0", n)
	}
	u := km.Lock("a")
	u()
}
