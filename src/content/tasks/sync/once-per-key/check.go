package main

func TestOnceMapOncePerKey(t *testing.T) {
	var m OnceMap[string, int]
	var calls sync.Map
	init := func(k string) (int, error) {
		n, _ := calls.LoadOrStore(k, new(atomic.Int32))
		n.(*atomic.Int32).Add(1)
		time.Sleep(5 * time.Millisecond)
		return len(k), nil
	}
	keys := []string{"a", "bb", "ccc"}
	var wg sync.WaitGroup
	for i := range 60 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			k := keys[i%3]
			if v, err := m.Get(k, init); err != nil || v != len(k) {
				t.Errorf("Get(%q) = %d, %v, ожидали %d", k, v, err, len(k))
			}
		}()
	}
	wg.Wait()
	for _, k := range keys {
		n, _ := calls.Load(k)
		if n == nil || n.(*atomic.Int32).Load() != 1 {
			t.Fatalf("init(%q) вызван %v раз, ожидали 1", k, n)
		}
	}
	if m.Len() != 3 {
		t.Fatalf("Len() = %d, ожидали 3", m.Len())
	}
}

func TestOnceMapErrorRemembered(t *testing.T) {
	var m OnceMap[int, string]
	boom := errors.New("шард недоступен")
	calls := 0
	init := func(int) (string, error) { calls++; return "", boom }
	for range 3 {
		if _, err := m.Get(7, init); !errors.Is(err, boom) {
			t.Fatalf("Get вернул %v, ожидали ошибку init", err)
		}
	}
	if calls != 1 {
		t.Fatalf("init для ключа с ошибкой вызван %d раз, ожидали 1", calls)
	}
}

func TestOnceMapKeysIndependent(t *testing.T) {
	var m OnceMap[string, int]
	release := make(chan struct{})
	started := make(chan struct{})
	go m.Get("медленный", func(string) (int, error) { close(started); <-release; return 1, nil })
	<-started
	done := make(chan int)
	go func() {
		v, _ := m.Get("быстрый", func(string) (int, error) { return 2, nil })
		done <- v
	}()
	select {
	case v := <-done:
		if v != 2 {
			t.Fatalf("Get(быстрый) = %d, ожидали 2", v)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Get другого ключа ждёт медленный init — init выполняется под общим мьютексом?")
	}
	close(release)
}
