package main

func chkWaitTimeout(t *testing.T, e *KeyedExecutor) {
	t.Helper()
	done := make(chan struct{})
	go func() { e.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Wait не дождался задач")
	}
}

func TestKeyedOrderPerKey(t *testing.T) {
	e := NewKeyedExecutor()
	var mu sync.Mutex
	got := map[string][]int{}
	var active [3]atomic.Int32
	for i := range 60 {
		k := i % 3
		key := fmt.Sprint("user-", k)
		e.Submit(key, func() {
			if active[k].Add(1) != 1 {
				t.Errorf("две задачи ключа %s выполнялись одновременно", key)
			}
			if i%7 == 0 {
				time.Sleep(time.Millisecond)
			}
			mu.Lock()
			got[key] = append(got[key], i)
			mu.Unlock()
			active[k].Add(-1)
		})
	}
	chkWaitTimeout(t, e)
	for k := range 3 {
		key := fmt.Sprint("user-", k)
		var want []int
		for i := k; i < 60; i += 3 {
			want = append(want, i)
		}
		if !reflect.DeepEqual(got[key], want) {
			t.Fatalf("задачи %s выполнены в порядке %v, ожидали %v", key, got[key], want)
		}
	}
}

func TestKeyedDifferentKeysParallel(t *testing.T) {
	e := NewKeyedExecutor()
	bRan := make(chan struct{})
	var aOK atomic.Bool
	e.Submit("a", func() {
		select {
		case <-bRan:
			aOK.Store(true)
		case <-time.After(time.Second):
		}
	})
	e.Submit("b", func() { close(bRan) })
	chkWaitTimeout(t, e)
	if !aOK.Load() {
		t.Fatal("задача ключа b не запустилась, пока работала задача ключа a")
	}
}

func TestKeyedSubmitDoesNotBlock(t *testing.T) {
	e := NewKeyedExecutor()
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		e.Submit("k", func() { <-release })
		for range 100 {
			e.Submit("k", func() {})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Submit ждёт, пока выполнится предыдущая задача ключа")
	}
	close(release)
	chkWaitTimeout(t, e)
}

func TestKeyedCleansUp(t *testing.T) {
	base := runtime.NumGoroutine()
	e := NewKeyedExecutor()
	var n atomic.Int32
	for i := range 1000 {
		e.Submit(fmt.Sprint("once-", i), func() { n.Add(1) })
	}
	chkWaitTimeout(t, e)
	if n.Load() != 1000 {
		t.Fatalf("выполнено %d задач из 1000", n.Load())
	}
	if k := e.Keys(); k != 0 {
		t.Fatalf("после выполнения всех задач Keys() = %d, ожидали 0 — записи ключей копятся", k)
	}
	for i := 0; i < 100 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if d := runtime.NumGoroutine() - base; d > 0 {
		t.Fatalf("после выполнения всех задач осталось %d горутин", d)
	}
}
