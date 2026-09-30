package main

func TestVersionWaitersWoken(t *testing.T) {
	v := NewVersion()
	results := make(chan int64, 3)
	for _, want := range []int64{3, 5, 10} {
		go func() {
			if err := v.WaitFor(context.Background(), want); err != nil {
				t.Errorf("WaitFor(%d) = %v", want, err)
			}
			results <- want
		}()
	}
	time.Sleep(20 * time.Millisecond)
	v.Set(5)
	got := map[int64]bool{}
	for range 2 {
		select {
		case r := <-results:
			got[r] = true
		case <-time.After(time.Second):
			t.Fatalf("после Set(5) проснулись только %v, ожидали ждущих 3 и 5 — Signal вместо Broadcast?", got)
		}
	}
	if !got[3] || !got[5] {
		t.Fatalf("после Set(5) проснулись %v, ожидали 3 и 5", got)
	}
	select {
	case r := <-results:
		t.Fatalf("ждущий версии %d проснулся при версии 5", r)
	case <-time.After(20 * time.Millisecond):
	}
	v.Set(4)
	if g := v.Get(); g != 5 {
		t.Fatalf("после Set(4) версия %d, ожидали 5 — версия не должна уменьшаться", g)
	}
	v.Set(10)
	select {
	case <-results:
	case <-time.After(time.Second):
		t.Fatal("ждущий версии 10 не проснулся после Set(10)")
	}
}

func TestVersionCancel(t *testing.T) {
	v := NewVersion()
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- v.WaitFor(ctx, 100) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("WaitFor после отмены = %v, ожидали context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("отмена контекста не разбудила WaitFor")
	}
	v.Set(1)
	if err := v.WaitFor(ctx, 1); err != nil {
		t.Fatalf("версия уже достигнута, а WaitFor с отменённым ctx вернул %v, ожидали nil", err)
	}
	if err := v.WaitFor(ctx, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("WaitFor недостигнутой версии с уже отменённым ctx = %v, ожидали context.Canceled", err)
	}
}

func TestVersionNoLeak(t *testing.T) {
	v := NewVersion()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	before := runtime.NumGoroutine()
	for i := range 200 {
		done := make(chan struct{})
		go func() { v.WaitFor(ctx, int64(i+1)); close(done) }()
		v.Set(int64(i + 1))
		<-done
	}
	time.Sleep(20 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+5 {
		t.Fatalf("после 200 успешных WaitFor с живым ctx горутин %d, было %d — что-то остаётся ждать отмены ctx", after, before)
	}
}
