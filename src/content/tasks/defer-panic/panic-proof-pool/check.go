package main

func chkMapDone[R any](t *testing.T, f func() []Result[R]) []Result[R] {
	t.Helper()
	ch := make(chan []Result[R], 1)
	go func() { ch <- f() }()
	select {
	case r := <-ch:
		return r
	case <-time.After(3 * time.Second):
		t.Fatal("Map не вернулась за 3 секунды — похоже, воркеры умерли или дедлок")
		return nil
	}
}

func TestPoolOrderAndValues(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7}
	res := chkMapDone(t, func() []Result[int] {
		return Map(context.Background(), items, 3, func(_ context.Context, x int) (int, error) {
			time.Sleep(time.Duration(8-x) * time.Millisecond)
			return x * x, nil
		})
	})
	if len(res) != len(items) {
		t.Fatalf("len(results) = %d, ожидали %d", len(res), len(items))
	}
	for i, r := range res {
		if r.Err != nil || r.Value != items[i]*items[i] {
			t.Fatalf("results[%d] = %+v, ожидали {Value:%d Err:nil} — порядок должен совпадать с items", i, r, items[i]*items[i])
		}
	}
}

func TestPoolSurvivesPanics(t *testing.T) {
	boom := errors.New("битые данные")
	items := make([]int, 20)
	for i := range items {
		items[i] = i
	}
	res := chkMapDone(t, func() []Result[string] {
		return Map(context.Background(), items, 2, func(_ context.Context, x int) (string, error) {
			switch x % 5 {
			case 1:
				panic(fmt.Sprintf("плохой элемент %d", x))
			case 3:
				panic(boom)
			}
			return strconv.Itoa(x), nil
		})
	})
	for i, r := range res {
		switch i % 5 {
		case 1:
			if r.Err == nil || !strings.Contains(r.Err.Error(), fmt.Sprintf("плохой элемент %d", i)) {
				t.Fatalf("results[%d].Err = %v, ожидали ошибку с текстом паники", i, r.Err)
			}
		case 3:
			if !errors.Is(r.Err, boom) {
				t.Fatalf("results[%d].Err = %v, ожидали, что errors.Is найдёт ошибку паники", i, r.Err)
			}
		default:
			if r.Err != nil || r.Value != strconv.Itoa(i) {
				t.Fatalf("results[%d] = %+v: после паник воркеры должны продолжать работу", i, r)
			}
		}
	}
}

func TestPoolLimit(t *testing.T) {
	var cur, peak atomic.Int32
	items := make([]int, 30)
	chkMapDone(t, func() []Result[int] {
		return Map(context.Background(), items, 4, func(_ context.Context, x int) (int, error) {
			n := cur.Add(1)
			for {
				p := peak.Load()
				if n <= p || peak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			cur.Add(-1)
			if x == 0 {
				panic("ой")
			}
			return x, nil
		})
	})
	if p := peak.Load(); p > 4 {
		t.Fatalf("одновременно работало %d задач, лимит 4", p)
	}
	if p := peak.Load(); p < 2 {
		t.Fatalf("одновременно работало не больше %d задач — воркеры не параллельны", p)
	}
}

func TestPoolCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	items := []int{0, 1, 2, 3, 4, 5}
	res := chkMapDone(t, func() []Result[int] {
		return Map(ctx, items, 1, func(_ context.Context, x int) (int, error) {
			calls.Add(1)
			if x == 2 {
				cancel()
			}
			return x, nil
		})
	})
	if n := calls.Load(); n != 3 {
		t.Fatalf("fn вызвана %d раз, ожидали 3: после отмены новые элементы не запускаются", n)
	}
	for i := 3; i < len(items); i++ {
		if !errors.Is(res[i].Err, context.Canceled) {
			t.Fatalf("results[%d].Err = %v, ожидали context.Canceled", i, res[i].Err)
		}
	}
}

func TestPoolNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 5 {
		Map(context.Background(), []int{1, 2, 3}, 8, func(_ context.Context, x int) (int, error) {
			panic("x")
		})
	}
	Map(context.Background(), []int(nil), 0, func(_ context.Context, x int) (int, error) { return x, nil })
	time.Sleep(20 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+1 {
		t.Fatalf("горутин было %d, стало %d — Map оставляет горутины после возврата", before, after)
	}
}
