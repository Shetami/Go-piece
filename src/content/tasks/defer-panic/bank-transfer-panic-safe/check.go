package main

func chkDone(t *testing.T, what string, f func()) {
	t.Helper()
	ch := make(chan struct{})
	go func() { defer close(ch); f() }()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: не завершилось за 3 секунды — дедлок?", what)
	}
}

func TestTransferBasic(t *testing.T) {
	b := NewBank(map[int]int{1: 100, 2: 50}, nil)
	chkDone(t, "перевод", func() {
		if err := b.Transfer(1, 2, 30); err != nil {
			t.Errorf("Transfer(1, 2, 30) = %v", err)
		}
	})
	if b.Balance(1) != 70 || b.Balance(2) != 80 {
		t.Fatalf("балансы %d и %d, ожидали 70 и 80", b.Balance(1), b.Balance(2))
	}
}

func TestTransferErrors(t *testing.T) {
	b := NewBank(map[int]int{1: 100, 2: 50}, nil)
	cases := []struct {
		from, to, amount int
		want             error
	}{
		{1, 2, 101, ErrInsufficient},
		{1, 1, 10, ErrSameAccount},
		{1, 9, 10, ErrNoAccount},
		{1, 2, 0, ErrBadAmount},
		{1, 2, -5, ErrBadAmount},
	}
	for _, c := range cases {
		var err error
		chkDone(t, fmt.Sprintf("Transfer(%d, %d, %d)", c.from, c.to, c.amount), func() { err = b.Transfer(c.from, c.to, c.amount) })
		if !errors.Is(err, c.want) {
			t.Fatalf("Transfer(%d, %d, %d) = %v, ожидали %v", c.from, c.to, c.amount, err, c.want)
		}
	}
	if b.Balance(1) != 100 || b.Balance(2) != 50 {
		t.Fatalf("после ошибок балансы изменились: %d и %d", b.Balance(1), b.Balance(2))
	}
}

func TestTransferNoDeadlock(t *testing.T) {
	b := NewBank(map[int]int{1: 1000, 2: 1000, 3: 1000}, nil)
	chkDone(t, "встречные переводы", func() {
		var wg sync.WaitGroup
		for i := range 300 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				from, to := i%3+1, (i+1)%3+1
				if i%2 == 0 {
					from, to = to, from
				}
				b.Transfer(from, to, 1)
			}()
		}
		wg.Wait()
	})
	if sum := b.Balance(1) + b.Balance(2) + b.Balance(3); sum != 3000 {
		t.Fatalf("сумма на счетах %d, ожидали 3000 — деньги потерялись", sum)
	}
}

func TestTransferAuditPanic(t *testing.T) {
	b := NewBank(map[int]int{1: 100, 2: 50}, func(from, to, amount int) {
		if amount == 13 {
			panic("аудит недоступен")
		}
	})
	var rec any
	chkDone(t, "перевод с паникой аудита", func() {
		defer func() { rec = recover() }()
		b.Transfer(2, 1, 13)
	})
	if rec != "аудит недоступен" {
		t.Fatalf("паника аудита должна лететь дальше, recover() = %v", rec)
	}
	var b1, b2 int
	chkDone(t, "чтение баланса после паники", func() { b1, b2 = b.Balance(1), b.Balance(2) })
	if b1 != 100 || b2 != 50 {
		t.Fatalf("после паники аудита балансы %d и %d, ожидали откат к 100 и 50", b1, b2)
	}
	chkDone(t, "перевод после паники", func() {
		if err := b.Transfer(1, 2, 10); err != nil {
			t.Errorf("перевод после паники: %v", err)
		}
	})
}
