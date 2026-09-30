package main

type chkUserEvent struct{ ID int }
type chkUserCreated chkUserEvent
type chkUserDeleted chkUserEvent

func chkWithin(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { defer close(done); f() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: зависло — похоже, обработчик зовётся под замком", what)
	}
}

func TestBusTypedDelivery(t *testing.T) {
	var b Bus
	var log []string
	Subscribe(&b, func(e chkUserCreated) { log = append(log, fmt.Sprintf("created-1 %d", e.ID)) })
	Subscribe(&b, func(e chkUserDeleted) { log = append(log, fmt.Sprintf("deleted %d", e.ID)) })
	Subscribe(&b, func(e chkUserCreated) { log = append(log, fmt.Sprintf("created-2 %d", e.ID)) })
	Subscribe(&b, func(e any) { log = append(log, "any") })

	if n := Publish(&b, chkUserCreated{7}); n != 2 {
		t.Fatalf("Publish(UserCreated) вызвал %d обработчиков, ожидали 2", n)
	}
	if n := Publish(&b, chkUserEvent{8}); n != 0 {
		t.Fatalf("UserEvent — другой тип, хоть и с той же структурой; вызвано %d", n)
	}
	if n := Publish(&b, "строка"); n != 0 {
		t.Fatalf("на string никто не подписан, а вызвано %d", n)
	}
	want := []string{"created-1 7", "created-2 7"}
	if !reflect.DeepEqual(log, want) {
		t.Fatalf("журнал = %q, ожидали %q (ровно тип E, порядок подписки)", log, want)
	}
}

func TestBusUnsubscribe(t *testing.T) {
	var b Bus
	var got []string
	h := func(tag string) func(int) { return func(int) { got = append(got, tag) } }
	unA := Subscribe(&b, h("a"))
	Subscribe(&b, h("b"))
	Subscribe(&b, h("c"))
	unA()
	unA() // повторная отписка не должна снять чужой обработчик
	Publish(&b, 1)
	if !reflect.DeepEqual(got, []string{"b", "c"}) {
		t.Fatalf("после двойной отписки a вызваны %q, ожидали [b c]", got)
	}
}

func TestBusReentrant(t *testing.T) {
	var b Bus
	var got []string
	var unSelf func()
	unSelf = Subscribe(&b, func(e chkUserCreated) {
		got = append(got, "self")
		unSelf()
		Subscribe(&b, func(e chkUserCreated) { got = append(got, "late") })
		Publish(&b, chkUserDeleted{e.ID})
	})
	Subscribe(&b, func(e chkUserCreated) { got = append(got, "second") })
	Subscribe(&b, func(e chkUserDeleted) { got = append(got, "deleted") })
	chkWithin(t, "Publish изнутри обработчика", func() { Publish(&b, chkUserCreated{1}) })
	if !reflect.DeepEqual(got, []string{"self", "deleted", "second"}) {
		t.Fatalf("первый Publish: %q, ожидали [self deleted second] — отписка и подписка во время Publish действуют со следующего", got)
	}
	got = nil
	chkWithin(t, "второй Publish", func() { Publish(&b, chkUserCreated{2}) })
	if !reflect.DeepEqual(got, []string{"second", "late"}) {
		t.Fatalf("второй Publish: %q, ожидали [second late]", got)
	}
}

func TestBusConcurrent(t *testing.T) {
	var b Bus
	var calls atomic.Int64
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			un := Subscribe(&b, func(x int) { calls.Add(1) })
			for j := range 100 {
				Publish(&b, i*j)
			}
			un()
			Publish(&b, "done")
		}()
	}
	chkWithin(t, "параллельные Subscribe/Publish", wg.Wait)
	if calls.Load() < 800 {
		t.Fatalf("каждая горутина подписана на время своих 100 Publish, а вызовов всего %d", calls.Load())
	}
	if n := Publish(&b, 0); n != 0 {
		t.Fatalf("после всех отписок Publish вызвал %d обработчиков", n)
	}
}
