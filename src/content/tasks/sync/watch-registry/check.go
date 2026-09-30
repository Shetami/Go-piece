package main

func chkRecv(t *testing.T, ch <-chan string) (string, bool) {
	t.Helper()
	select {
	case v, ok := <-ch:
		return v, ok
	case <-time.After(time.Second):
		t.Fatal("уведомление не пришло")
	}
	return "", false
}

func TestRegistryWatchBasics(t *testing.T) {
	r := NewRegistry()
	r.Set("db", "10.0.0.1")
	ch, cancel := r.Watch("db")
	if v, _ := chkRecv(t, ch); v != "10.0.0.1" {
		t.Fatalf("первое уведомление = %q, ожидали текущее значение 10.0.0.1", v)
	}
	r.Set("db", "10.0.0.2")
	if v, _ := chkRecv(t, ch); v != "10.0.0.2" {
		t.Fatalf("уведомление = %q, ожидали 10.0.0.2", v)
	}
	r.Set("cache", "x")
	select {
	case v := <-ch:
		t.Fatalf("подписчик db получил %q от чужого ключа", v)
	default:
	}
	if g, ok := r.Get("db"); !ok || g != "10.0.0.2" {
		t.Fatalf("Get(db) = %q, %v", g, ok)
	}
	cancel()
	cancel()
	if _, ok := chkRecv(t, ch); ok {
		t.Fatal("после cancel канал должен быть закрыт")
	}
	if n := r.Watchers("db"); n != 0 {
		t.Fatalf("Watchers(db) = %d после отписки, ожидали 0", n)
	}
}

func TestRegistrySlowWatcherGetsLatest(t *testing.T) {
	r := NewRegistry()
	ch, cancel := r.Watch("flag")
	defer cancel()
	done := make(chan struct{})
	go func() {
		for i := range 100 {
			r.Set("flag", fmt.Sprint("v", i))
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Set заблокировался на подписчике, который не читает канал")
	}
	if v, _ := chkRecv(t, ch); v != "v99" {
		t.Fatalf("медленный подписчик получил %q, ожидали последнее значение v99", v)
	}
	select {
	case v := <-ch:
		t.Fatalf("после последнего значения в канале осталось ещё %q", v)
	default:
	}
}

func TestRegistryCancelDuringSet(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
					r.Set("k", fmt.Sprint(i))
				}
			}
		}()
	}
	for range 300 {
		ch, cancel := r.Watch("k")
		go cancel()
		cancel()
		for range ch {
		}
	}
	close(stop)
	wg.Wait()
	if n := r.Watchers("k"); n != 0 {
		t.Fatalf("Watchers(k) = %d после всех отписок", n)
	}
}
