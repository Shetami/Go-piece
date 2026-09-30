package main

func chkRecv[T any](t *testing.T, ch <-chan T, what string) (T, bool) {
	t.Helper()
	if ch == nil {
		t.Fatalf("%s: канал nil", what)
	}
	select {
	case v, ok := <-ch:
		return v, ok
	case <-time.After(time.Second):
		t.Fatalf("%s: ничего не пришло за секунду", what)
	}
	var zero T
	return zero, false
}

func TestHeartbeatResultsWithoutListener(t *testing.T) {
	jobs := make(chan int)
	hb, res := Work(context.Background(), time.Millisecond, jobs, func(x int) int { return x * 10 })
	go func() {
		for i := 1; i <= 5; i++ {
			jobs <- i
			time.Sleep(2 * time.Millisecond) // тикер успевает тикнуть, а heartbeat никто не читает
		}
		close(jobs)
	}()
	var got []int
	for {
		v, ok := chkRecv(t, res, "results")
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{10, 20, 30, 40, 50}) {
		t.Fatalf("результаты %v, ожидали [10 20 30 40 50]", got)
	}
	for {
		if _, ok := chkRecv(t, hb, "heartbeat после закрытия jobs"); !ok {
			break
		}
	}
}

func TestHeartbeatWhileIdle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	hb, res := Work(ctx, 5*time.Millisecond, make(chan int), func(x int) int { return x })
	for i := range 3 {
		if _, ok := chkRecv(t, hb, fmt.Sprintf("пульс %d в простое", i+1)); !ok {
			t.Fatal("heartbeat закрылся, хотя воркер работает")
		}
	}
	cancel()
	for {
		if _, ok := chkRecv(t, res, "results после отмены"); !ok {
			break
		}
	}
	for {
		if _, ok := chkRecv(t, hb, "heartbeat после отмены"); !ok {
			break
		}
	}
}

func TestHeartbeatWhileResultNotTaken(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan int, 1)
	jobs <- 7
	hb, res := Work(ctx, 5*time.Millisecond, jobs, func(x int) int { return x + 1 })
	time.Sleep(30 * time.Millisecond) // результат готов, но его не забирают
	for len(hb) > 0 {
		<-hb
	}
	if _, ok := chkRecv(t, hb, "пульс, пока результат никто не забирает"); !ok {
		t.Fatal("heartbeat закрылся раньше времени")
	}
	if v, _ := chkRecv(t, res, "results"); v != 8 {
		t.Fatalf("результат %d, ожидали 8", v)
	}
}
