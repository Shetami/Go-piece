package main

func chkRecv(t *testing.T, ch <-chan int, what string) int {
	t.Helper()
	select {
	case v, ok := <-ch:
		if !ok {
			t.Fatalf("%s: канал закрыт раньше времени", what)
		}
		return v
	case <-time.After(time.Second):
		t.Fatalf("%s: значение так и не пришло", what)
		return 0
	}
}

func TestLatestSetNeverBlocks(t *testing.T) {
	l := NewLatest[int]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l.Watch(ctx) // наблюдатель, который ничего не читает
	done := make(chan struct{})
	go func() {
		for i := range 1000 {
			l.Set(i)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Set заблокировался на наблюдателе, который не читает")
	}
}

func TestLatestCurrentAndNext(t *testing.T) {
	l := NewLatest[int]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l.Set(10)
	ch := l.Watch(ctx)
	if v := chkRecv(t, ch, "текущее значение"); v != 10 {
		t.Fatalf("новый наблюдатель получил %d, ожидали текущее 10", v)
	}
	select {
	case v := <-ch:
		t.Fatalf("без нового Set пришло %d — значение не должно повторяться", v)
	case <-time.After(30 * time.Millisecond):
	}
	l.Set(11)
	if v := chkRecv(t, ch, "после Set"); v != 11 {
		t.Fatalf("после Set(11) пришло %d", v)
	}
	fresh := NewLatest[int]().Watch(ctx)
	select {
	case v := <-fresh:
		t.Fatalf("значения ещё не было, а наблюдатель получил %d", v)
	case <-time.After(30 * time.Millisecond):
	}
}

func TestLatestSlowReaderGetsLatest(t *testing.T) {
	l := NewLatest[int]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := l.Watch(ctx)
	for i := 1; i <= 1000; i++ {
		l.Set(i)
	}
	var got []int
	for {
		v := chkRecv(t, ch, "медленный читатель")
		got = append(got, v)
		if v == 1000 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !slices.IsSorted(got) || len(slices.Compact(slices.Clone(got))) != len(got) {
		t.Fatalf("значения пришли не по возрастанию или с повторами: %v", got)
	}
	if len(got) > 5 {
		t.Fatalf("медленный читатель получил %d значений — промежуточные должны пропускаться, а не копиться", len(got))
	}
}

func TestLatestCancelCloses(t *testing.T) {
	base := runtime.NumGoroutine()
	l := NewLatest[int]()
	ctx, cancel := context.WithCancel(context.Background())
	chs := []<-chan int{l.Watch(ctx), l.Watch(ctx)}
	l.Set(1) // никто не читает: горутины висят на отправке
	time.Sleep(10 * time.Millisecond)
	cancel()
	for i, ch := range chs {
		deadline := time.After(time.Second)
	drain:
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					break drain
				}
			case <-deadline:
				t.Fatalf("канал наблюдателя %d не закрылся после отмены ctx", i)
			}
		}
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if d := runtime.NumGoroutine() - base; d > 0 {
		t.Fatalf("после отмены осталось %d горутин наблюдателей", d)
	}
}
