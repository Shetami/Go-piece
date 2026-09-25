package main

func TestBarrierBlocksUntilAll(t *testing.T) {
	b := NewBarrier(3)
	var passed atomic.Int32
	for range 2 {
		go func() {
			b.Wait()
			passed.Add(1)
		}()
	}
	time.Sleep(50 * time.Millisecond)
	if p := passed.Load(); p != 0 {
		t.Fatalf("пришли двое из трёх, а прошло %d", p)
	}
	done := make(chan struct{})
	go func() { b.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("третий участник не прошёл барьер")
	}
	deadline := time.Now().Add(time.Second)
	for passed.Load() != 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if p := passed.Load(); p != 2 {
		t.Fatalf("после прихода третьего прошли не все: %d из 2 ждавших", p)
	}
}

func TestBarrierReusable(t *testing.T) {
	const n, rounds = 4, 50
	b := NewBarrier(n)
	var counts [rounds]atomic.Int32
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range rounds {
				counts[r].Add(1)
				b.Wait()
				// За барьером круга r все n уже отметились в нём.
				if c := counts[r].Load(); c != n {
					t.Errorf("круг %d: за барьером видно %d отметок из %d", r, c, n)
					return
				}
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("многоразовый барьер завис")
	}
}

func TestBarrierOne(t *testing.T) {
	b := NewBarrier(1)
	done := make(chan struct{})
	go func() { b.Wait(); b.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("барьер на одного должен пропускать сразу")
	}
}
