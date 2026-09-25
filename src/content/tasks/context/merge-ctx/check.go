package main

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(time.Second):
		return false
	}
}

func TestMergeCancelA(t *testing.T) {
	a, cancelA := context.WithCancel(context.Background())
	m, cancel := Merge(a, context.Background())
	defer cancel()
	cancelA()
	if !isDone(m) || !errors.Is(m.Err(), context.Canceled) {
		t.Fatalf("отмена a должна отменить результат")
	}
}

func TestMergeCancelB(t *testing.T) {
	b, cancelB := context.WithCancelCause(context.Background())
	m, cancel := Merge(context.Background(), b)
	defer cancel()
	why := errors.New("сервис останавливается")
	cancelB(why)
	if !isDone(m) {
		t.Fatalf("отмена b должна отменить результат")
	}
	if !errors.Is(context.Cause(m), why) {
		t.Fatalf("Cause = %v, ожидали причину отмены b", context.Cause(m))
	}
}

func TestMergeDeadlineB(t *testing.T) {
	b, cancelB := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelB()
	m, cancel := Merge(context.Background(), b)
	defer cancel()
	if !isDone(m) || !errors.Is(context.Cause(m), context.DeadlineExceeded) {
		t.Fatalf("истёкший b: Cause = %v, ожидали DeadlineExceeded", context.Cause(m))
	}
}

func TestMergeValuesFromA(t *testing.T) {
	type key struct{}
	a := context.WithValue(context.Background(), key{}, "запрос-1")
	m, cancel := Merge(a, context.Background())
	defer cancel()
	if m.Value(key{}) != "запрос-1" {
		t.Fatalf("значения должны браться из a")
	}
	if m.Err() != nil {
		t.Fatalf("ни a, ни b не отменены, а результат уже: %v", m.Err())
	}
}

func TestMergeCancelFunc(t *testing.T) {
	m, cancel := Merge(context.Background(), context.Background())
	cancel()
	cancel()
	if !isDone(m) || !errors.Is(m.Err(), context.Canceled) {
		t.Fatalf("cancel должна отменять результат")
	}
}

func TestMergeNoLeak(t *testing.T) {
	b, cancelB := context.WithCancel(context.Background())
	defer cancelB()
	before := runtime.NumGoroutine()
	for range 100 {
		_, cancel := Merge(context.Background(), b)
		cancel()
	}
	time.Sleep(20 * time.Millisecond)
	if n := runtime.NumGoroutine(); n > before+2 {
		t.Fatalf("после 100 Merge+cancel горутин стало больше на %d", n-before)
	}
}
