package main

func TestStoreMaxSequential(t *testing.T) {
	var p atomic.Int64
	p.Store(10)
	if StoreMax(&p, 5) || p.Load() != 10 {
		t.Fatalf("меньшее значение не должно записываться: %d", p.Load())
	}
	if StoreMax(&p, 10) {
		t.Fatalf("равное значение — не обновление")
	}
	if !StoreMax(&p, 42) || p.Load() != 42 {
		t.Fatalf("большее значение должно записаться: %d", p.Load())
	}
	var neg atomic.Int64
	neg.Store(-100)
	if !StoreMax(&neg, -5) || neg.Load() != -5 {
		t.Fatalf("отрицательные числа: %d", neg.Load())
	}
}

func TestStoreMaxConcurrent(t *testing.T) {
	for range 20 {
		var p atomic.Int64
		var updates atomic.Int64
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range 1000 {
					if StoreMax(&p, int64(i*8+g)) {
						updates.Add(1)
					}
				}
			}()
		}
		wg.Wait()
		if got := p.Load(); got != 7999 {
			t.Fatalf("после гонки максимум %d, ожидали 7999 — какая-то запись затёрла большее значение", got)
		}
		if updates.Load() < 1 {
			t.Fatalf("ни одного обновления")
		}
	}
}
