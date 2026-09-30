package main

func chkDone(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { f(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("зависание: %s", what)
	}
}

func TestQueueFIFOAndBlocking(t *testing.T) {
	q := NewQueue[int](2)
	q.Put(1)
	q.Put(2)
	put3 := make(chan error, 1)
	go func() { put3 <- q.Put(3) }()
	select {
	case <-put3:
		t.Fatal("Put в полную очередь не заблокировался")
	case <-time.After(30 * time.Millisecond):
	}
	for want := 1; want <= 3; want++ {
		var v int
		var ok bool
		chkDone(t, "Take", func() { v, ok = q.Take() })
		if !ok || v != want {
			t.Fatalf("Take = %d, %v, ожидали %d, true (FIFO)", v, ok, want)
		}
	}
	if err := <-put3; err != nil {
		t.Fatalf("Put после освобождения места = %v", err)
	}
	if q.Len() != 0 {
		t.Fatalf("Len() = %d, ожидали 0", q.Len())
	}
}

func TestQueueCloseWakesAll(t *testing.T) {
	q := NewQueue[string](1)
	var takers sync.WaitGroup
	for range 3 {
		takers.Add(1)
		go func() {
			defer takers.Done()
			if v, ok := q.Take(); ok {
				t.Errorf("Take из пустой закрытой очереди = %q, true", v)
			}
		}()
	}
	time.Sleep(30 * time.Millisecond)
	q.Close()
	q.Close()
	chkDone(t, "после Close проснулись не все ждущие Take — Signal вместо Broadcast?", takers.Wait)

	full := NewQueue[string](1)
	full.Put("x")
	errs := make(chan error, 3)
	for range 3 {
		go func() { errs <- full.Put("y") }()
	}
	time.Sleep(30 * time.Millisecond)
	full.Close()
	for range 3 {
		select {
		case err := <-errs:
			if !errors.Is(err, ErrClosed) {
				t.Fatalf("Put, ждавший места в закрытой очереди, вернул %v, ожидали ErrClosed", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Put, ждавший места, не проснулся после Close")
		}
	}
	if err := full.Put("z"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Put после Close = %v, ожидали ErrClosed", err)
	}
	if v, ok := full.Take(); !ok || v != "x" {
		t.Fatalf("после Close Take = %q, %v, ожидали оставшийся x", v, ok)
	}
}

func TestQueueDrainAfterClose(t *testing.T) {
	q := NewQueue[int](5)
	q.Put(1)
	q.Put(2)
	q.Close()
	a, ok1 := q.Take()
	b, ok2 := q.Take()
	_, ok3 := q.Take()
	if a != 1 || b != 2 || !ok1 || !ok2 || ok3 {
		t.Fatalf("после Close: (%d %v) (%d %v) (_ %v), ожидали остаток 1, 2 и затем false", a, ok1, b, ok2, ok3)
	}
}

func TestQueueManyProducersConsumers(t *testing.T) {
	q := NewQueue[int](2)
	var prod, cons sync.WaitGroup
	var sum, count atomic.Int64
	for range 4 {
		cons.Add(1)
		go func() {
			defer cons.Done()
			for {
				v, ok := q.Take()
				if !ok {
					return
				}
				sum.Add(int64(v))
				count.Add(1)
			}
		}()
	}
	for p := range 4 {
		prod.Add(1)
		go func() {
			defer prod.Done()
			for i := range 1000 {
				q.Put(p*1000 + i)
			}
		}()
	}
	chkDone(t, "4 производителя и 4 потребителя на очереди из 2 мест — потерянная побудка?", func() {
		prod.Wait()
		q.Close()
		cons.Wait()
	})
	if count.Load() != 4000 || sum.Load() != 3998*4000/2+2000 {
		t.Fatalf("потребители получили %d элементов с суммой %d, ожидали 4000 и %d", count.Load(), sum.Load(), 3998*4000/2+2000)
	}
}
