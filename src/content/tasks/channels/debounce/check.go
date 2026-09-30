package main

func chkNextVal(t *testing.T, out <-chan int, within time.Duration, what string) (int, bool) {
	t.Helper()
	if out == nil {
		t.Fatal("Debounce вернул nil-канал")
	}
	select {
	case v, ok := <-out:
		return v, ok
	case <-time.After(within):
		t.Fatalf("%s: ничего не пришло за %v", what, within)
	}
	return 0, false
}

func chkSend(t *testing.T, in chan<- int, v int) {
	t.Helper()
	select {
	case in <- v:
	case <-time.After(time.Second):
		t.Fatalf("Debounce не забрал значение %d из входа", v)
	}
}

func TestDebounceBurst(t *testing.T) {
	in := make(chan int)
	out := Debounce(context.Background(), in, 30*time.Millisecond, 0)
	chkSend(t, in, 1)
	chkSend(t, in, 2)
	chkSend(t, in, 3)
	if v, ok := chkNextVal(t, out, time.Second, "после серии 1,2,3"); !ok || v != 3 {
		t.Fatalf("после серии пришло %v (ok=%v), ожидали только последнее — 3", v, ok)
	}
	select {
	case v := <-out:
		t.Fatalf("после серии пришло лишнее событие %v", v)
	case <-time.After(60 * time.Millisecond):
	}
	chkSend(t, in, 4)
	if v, _ := chkNextVal(t, out, time.Second, "новая серия"); v != 4 {
		t.Fatalf("новая серия: пришло %v, ожидали 4", v)
	}
	close(in)
	if _, ok := chkNextVal(t, out, time.Second, "закрытие"); ok {
		t.Fatal("вход закрыт без ожидающих событий, ожидали закрытие выхода")
	}
}

func TestDebounceFlushOnClose(t *testing.T) {
	in := make(chan int)
	out := Debounce(context.Background(), in, time.Hour, 0)
	chkSend(t, in, 5)
	chkSend(t, in, 6)
	close(in)
	v, ok := chkNextVal(t, out, time.Second, "вход закрыт с ожидающим событием")
	if !ok || v != 6 {
		t.Fatalf("при закрытии входа пришло %v (ok=%v), ожидали 6 сразу, не дожидаясь wait", v, ok)
	}
	if _, ok := chkNextVal(t, out, time.Second, "после хвоста"); ok {
		t.Fatal("после хвоста выход должен закрыться")
	}
}

func TestDebounceMaxWait(t *testing.T) {
	in := make(chan int)
	out := Debounce(context.Background(), in, 40*time.Millisecond, 100*time.Millisecond)
	var got atomic.Int64
	var lastSeen atomic.Int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		for v := range out {
			got.Add(1)
			lastSeen.Store(int64(v))
		}
	}()
	for i := 1; i <= 40; i++ {
		chkSend(t, in, i)
		time.Sleep(10 * time.Millisecond)
	}
	during := got.Load()
	close(in)
	<-done
	if during < 2 {
		t.Fatalf("события шли непрерывно 400 мс при maxWait 100 мс, а наружу за это время ушло %d — ожидали хотя бы 2", during)
	}
	if lastSeen.Load() != 40 {
		t.Fatalf("последним ушло %d, ожидали 40 (хвост при закрытии)", lastSeen.Load())
	}
}

func TestDebounceCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int)
	out := Debounce(ctx, in, time.Hour, 0)
	chkSend(t, in, 1)
	cancel()
	if v, ok := chkNextVal(t, out, time.Second, "после отмены"); ok {
		t.Fatalf("после отмены пришло %v, ожидали закрытие выхода", v)
	}
}
