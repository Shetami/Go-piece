package main

func TestBarrierRounds(t *testing.T) {
	const n, rounds = 5, 100
	b := NewBarrier(n)
	var arrived [rounds]atomic.Int32
	var leaders [rounds]atomic.Int32
	var wg sync.WaitGroup
	for g := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range rounds {
				if (g+r)%3 == 0 {
					time.Sleep(time.Duration(g) * 50 * time.Microsecond)
				}
				arrived[r].Add(1)
				last, err := b.Wait(context.Background())
				if err != nil {
					t.Errorf("Wait = %v", err)
					return
				}
				if last {
					leaders[r].Add(1)
				}
				if a := arrived[r].Load(); a != n {
					t.Errorf("раунд %d: прошли барьер, когда пришло %d из %d", r, a, n)
					return
				}
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("барьер завис при повторном использовании")
	}
	for r := range rounds {
		if l := leaders[r].Load(); l != 1 {
			t.Fatalf("раунд %d: true получили %d горутин, ожидали ровно одну", r, l)
		}
	}
}

func TestBarrierSingle(t *testing.T) {
	b := NewBarrier(1)
	for range 3 {
		if last, err := b.Wait(context.Background()); !last || err != nil {
			t.Fatalf("NewBarrier(1).Wait = %v, %v; ожидали true, nil без ожидания", last, err)
		}
	}
}

func TestBarrierCancelLeaves(t *testing.T) {
	b := NewBarrier(2)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	res := make(chan error, 1)
	go func() { _, err := b.Wait(ctx); res <- err }()
	select {
	case err := <-res:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Wait с истёкшим ctx = %v, ожидали DeadlineExceeded", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Wait не вернулся по истечении ctx")
	}
	// Ушедший не считается: одна новая горутина не должна пройти барьер на двоих.
	single := make(chan struct{})
	go func() { b.Wait(context.Background()); close(single) }()
	select {
	case <-single:
		t.Fatal("барьер на 2 пропустил одну горутину — ушедшая по отмене всё ещё засчитана")
	case <-time.After(50 * time.Millisecond):
	}
	b.Wait(context.Background())
	select {
	case <-single:
	case <-time.After(time.Second):
		t.Fatal("вторая горутина пришла, а первая так и не прошла барьер")
	}
}
