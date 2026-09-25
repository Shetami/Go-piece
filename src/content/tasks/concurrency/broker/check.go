package main

func TestBrokerFanOut(t *testing.T) {
	b := NewBroker()
	a, _ := b.Subscribe(4)
	c, _ := b.Subscribe(4)
	if n := b.Publish("привет"); n != 2 {
		t.Fatalf("Publish доставил %d, ожидали 2", n)
	}
	if m := <-a; m != "привет" {
		t.Fatalf("первый подписчик получил %q", m)
	}
	if m := <-c; m != "привет" {
		t.Fatalf("второй подписчик получил %q", m)
	}
}

func TestBrokerNoSubscribers(t *testing.T) {
	if n := NewBroker().Publish("x"); n != 0 {
		t.Fatalf("без подписчиков доставлено %d", n)
	}
}

func TestBrokerUnsubscribe(t *testing.T) {
	b := NewBroker()
	ch, unsub := b.Subscribe(4)
	unsub()
	unsub() // повторная отписка не должна паниковать
	if _, ok := <-ch; ok {
		t.Fatalf("после отписки канал должен быть закрыт")
	}
	if n := b.Publish("x"); n != 0 {
		t.Fatalf("после отписки доставлено %d", n)
	}
}

func TestBrokerSlowSubscriber(t *testing.T) {
	b := NewBroker()
	slow, _ := b.Subscribe(1)
	fast, _ := b.Subscribe(10)
	done := make(chan struct{})
	go func() {
		for i := range 5 {
			b.Publish(fmt.Sprint(i))
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("Publish заблокировался на подписчике с полным буфером")
	}
	if len(slow) != 1 || len(fast) != 5 {
		t.Fatalf("в буферах %d и %d сообщений, ожидали 1 и 5", len(slow), len(fast))
	}
}

func TestBrokerConcurrent(t *testing.T) {
	b := NewBroker()
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			ch, unsub := b.Subscribe(1)
			b.Publish("x")
			unsub()
			for range ch {
			}
		}()
		go func() {
			defer wg.Done()
			b.Publish("y")
		}()
	}
	wg.Wait()
}
