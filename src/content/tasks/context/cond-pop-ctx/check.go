package main

func chkPop[T any](q *Queue[T], ctx context.Context) (T, error, bool) {
	type res struct {
		v   T
		err error
	}
	ch := make(chan res, 1)
	go func() { v, err := q.Pop(ctx); ch <- res{v, err} }()
	select {
	case r := <-ch:
		return r.v, r.err, true
	case <-time.After(2 * time.Second):
		var zero T
		return zero, nil, false
	}
}

func TestPopWaitsForPush(t *testing.T) {
	q := NewQueue[string]()
	go func() { time.Sleep(20 * time.Millisecond); q.Push("a"); q.Push("b") }()
	v, err, ok := chkPop(q, context.Background())
	if !ok || err != nil || v != "a" {
		t.Fatalf("Pop = %q, %v (вернулся: %v); ожидали «a» после Push", v, err, ok)
	}
	if v, _, _ := chkPop(q, context.Background()); v != "b" {
		t.Fatalf("второй Pop = %q, ожидали «b» — порядок FIFO", v)
	}
}

func TestPopCancelWhileWaiting(t *testing.T) {
	why := errors.New("воркер останавливается")
	q := NewQueue[int]()
	ctx, cancel := context.WithCancelCause(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel(why) }()
	_, err, ok := chkPop(q, ctx)
	if !ok {
		t.Fatalf("Pop на пустой очереди не заметил отмену контекста и висит")
	}
	if !errors.Is(err, why) {
		t.Fatalf("err = %v, ожидали причину отмены", err)
	}
}

func TestPopPreCanceledKeepsItem(t *testing.T) {
	q := NewQueue[int]()
	q.Push(7)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range 20 {
		if _, err, _ := chkPop(q, ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("ctx отменён заранее, а Pop вернул err = %v", err)
		}
	}
	if q.Len() != 1 {
		t.Fatalf("отменённый Pop забрал элемент — он потерян")
	}
}

func TestPopManyWaiters(t *testing.T) {
	q := NewQueue[int]()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	var got, canceled atomic.Int32
	for i := range 20 {
		c := context.Background()
		if i%2 == 0 {
			c = ctx
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := q.Pop(c); err == nil {
				got.Add(1)
			} else {
				canceled.Add(1)
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	cancel()
	time.Sleep(20 * time.Millisecond)
	for i := range 10 {
		q.Push(i)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("не все ожидающие вернулись: получили %d, отменено %d", got.Load(), canceled.Load())
	}
	if got.Load() != 10 || canceled.Load() != 10 || q.Len() != 0 {
		t.Fatalf("получили %d, отменено %d, осталось %d; ожидали 10, 10, 0", got.Load(), canceled.Load(), q.Len())
	}
}

func TestPopNoGoroutineLeak(t *testing.T) {
	q := NewQueue[int]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	before := runtime.NumGoroutine()
	for i := range 200 {
		q.Push(i)
		if _, err := q.Pop(ctx); err != nil {
			t.Fatalf("Pop: %v", err)
		}
	}
	time.Sleep(20 * time.Millisecond)
	if n := runtime.NumGoroutine(); n > before+5 {
		t.Fatalf("после 200 успешных Pop горутин стало больше на %d — кто-то ждёт отмены вечно", n-before)
	}
}
