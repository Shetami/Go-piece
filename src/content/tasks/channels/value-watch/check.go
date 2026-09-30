package main

func TestVarBroadcast(t *testing.T) {
	x := NewVar("v1")
	v, ch := x.Get()
	if v != "v1" || ch == nil {
		t.Fatalf("Get = %q, канал %v; ожидали \"v1\" и не-nil канал", v, ch)
	}
	var woke atomic.Int64
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ch:
				woke.Add(1)
			case <-time.After(time.Second):
			}
		}()
	}
	time.Sleep(10 * time.Millisecond)
	x.Set("v2")
	wg.Wait()
	if woke.Load() != 10 {
		t.Fatalf("после Set проснулись %d из 10 ожидающих", woke.Load())
	}
	v2, ch2 := x.Get()
	if v2 != "v2" {
		t.Fatalf("Get после Set = %q, ожидали v2", v2)
	}
	select {
	case <-ch2:
		t.Fatal("канал из Get после Set уже закрыт — ожидание следующего изменения не работает")
	default:
	}
	x.Set("v3") // второй Set не должен паниковать
	select {
	case <-ch2:
	case <-time.After(time.Second):
		t.Fatal("второй Set не закрыл канал нового раунда")
	}
}

func TestVarWaitFor(t *testing.T) {
	x := NewVar(0)
	go func() {
		for i := 1; i <= 100; i++ {
			x.Set(i)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	v, err := x.WaitFor(ctx, func(n int) bool { return n >= 50 })
	if err != nil || v < 50 {
		t.Fatalf("WaitFor(>=50) = %d, %v; ожидали значение >= 50 и nil", v, err)
	}
	if v, err := x.WaitFor(ctx, func(n int) bool { return n >= 0 }); err != nil || v < 0 {
		t.Fatalf("текущее значение уже подходит, а WaitFor = %d, %v", v, err)
	}
}

func TestVarNoLostWakeup(t *testing.T) {
	for i := range 300 {
		x := NewVar(false)
		res := make(chan error, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := x.WaitFor(ctx, func(b bool) bool { return b })
			res <- err
		}()
		x.Set(true)
		if err := <-res; err != nil {
			t.Fatalf("попытка %d: Set(true) случился, а WaitFor его пропустил и вернул %v", i, err)
		}
	}
}

func TestVarWaitForCancel(t *testing.T) {
	x := NewVar(3)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	v, err := x.WaitFor(ctx, func(n int) bool { return n > 10 })
	if !errors.Is(err, context.DeadlineExceeded) || v != 3 {
		t.Fatalf("WaitFor по таймауту = %d, %v; ожидали 3, DeadlineExceeded", v, err)
	}
}
