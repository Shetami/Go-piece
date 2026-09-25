package main

func TestWithLockRuns(t *testing.T) {
	var mu sync.Mutex
	ran := false
	WithLock(&mu, func() {
		ran = true
		if mu.TryLock() {
			t.Fatalf("во время f мьютекс должен быть захвачен")
		}
	})
	if !ran {
		t.Fatalf("f не вызвана")
	}
	if !mu.TryLock() {
		t.Fatalf("после WithLock мьютекс остался захвачен")
	}
}

func TestWithLockPanic(t *testing.T) {
	var mu sync.Mutex
	func() {
		defer func() {
			if r := recover(); r != "бум" {
				t.Fatalf("паника должна пролететь наружу как есть, получили %v", r)
			}
		}()
		WithLock(&mu, func() { panic("бум") })
	}()
	if !mu.TryLock() {
		t.Fatalf("после паники в f мьютекс остался захвачен")
	}
}

func TestWithLockConcurrent(t *testing.T) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	n := 0
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WithLock(&mu, func() { n++ })
		}()
	}
	wg.Wait()
	if n != 100 {
		t.Fatalf("n = %d, ожидали 100", n)
	}
}
