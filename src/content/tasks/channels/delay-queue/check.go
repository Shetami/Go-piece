package main

func chkPop(t *testing.T, q *DelayQueue[string], within time.Duration, what string) string {
	t.Helper()
	res := make(chan string, 1)
	go func() {
		v, _ := q.Pop(context.Background())
		res <- v
	}()
	select {
	case v := <-res:
		return v
	case <-time.After(within):
		t.Fatalf("%s: Pop не вернулся за %v", what, within)
	}
	return ""
}

func TestDelayQueueOrder(t *testing.T) {
	q := NewDelayQueue[string]()
	q.Put("c", 30*time.Millisecond)
	q.Put("a", 10*time.Millisecond)
	q.Put("b", 20*time.Millisecond)
	q.Put("x", 20*time.Millisecond)
	if n := q.Len(); n != 4 {
		t.Fatalf("Len = %d, ожидали 4", n)
	}
	var got []string
	for range 4 {
		got = append(got, chkPop(t, q, 2*time.Second, "по порядку готовности"))
	}
	if !reflect.DeepEqual(got, []string{"a", "b", "x", "c"}) {
		t.Fatalf("получили %v, ожидали [a b x c] (по времени готовности, равные — по порядку Put)", got)
	}
}

func TestDelayQueueNotEarly(t *testing.T) {
	q := NewDelayQueue[string]()
	q.Put("later", 50*time.Millisecond)
	start := time.Now()
	chkPop(t, q, 2*time.Second, "отложенный элемент")
	if el := time.Since(start); el < 40*time.Millisecond {
		t.Fatalf("элемент с задержкой 50 мс выдан через %v — раньше срока", el)
	}
}

func TestDelayQueueEarlierArrives(t *testing.T) {
	q := NewDelayQueue[string]()
	q.Put("hour", time.Hour)
	res := make(chan string, 1)
	go func() {
		v, _ := q.Pop(context.Background())
		res <- v
	}()
	time.Sleep(10 * time.Millisecond) // Pop уже ждёт элемент через час
	q.Put("soon", 10*time.Millisecond)
	select {
	case v := <-res:
		if v != "soon" {
			t.Fatalf("получили %q, ожидали soon", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Pop ждал элемент через час и не заметил добавленный более ранний")
	}
}

func TestDelayQueueManyConsumers(t *testing.T) {
	q := NewDelayQueue[string]()
	res := make(chan string, 3)
	for range 3 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			v, _ := q.Pop(ctx)
			res <- v
		}()
	}
	time.Sleep(10 * time.Millisecond) // все трое ждут на пустой очереди
	q.Put("1", 0)
	q.Put("2", 0)
	q.Put("3", 0)
	var got []string
	for range 3 {
		select {
		case v := <-res:
			got = append(got, v)
		case <-time.After(time.Second):
			t.Fatalf("три готовых элемента и три потребителя, а получено только %v", got)
		}
	}
	slices.Sort(got)
	if !reflect.DeepEqual(got, []string{"1", "2", "3"}) {
		t.Fatalf("потребители получили %v, ожидали по одному из 1, 2, 3", got)
	}
}

func TestDelayQueueCancel(t *testing.T) {
	q := NewDelayQueue[string]()
	q.Put("hour", time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	v, err := q.Pop(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || v != "" {
		t.Fatalf("Pop по таймауту = %q, %v; ожидали \"\", DeadlineExceeded", v, err)
	}
	if q.Len() != 1 {
		t.Fatalf("после отменённого Pop Len = %d, элемент не должен пропасть", q.Len())
	}
}
