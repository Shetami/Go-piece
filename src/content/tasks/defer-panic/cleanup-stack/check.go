package main

func TestCleanupReverseOrder(t *testing.T) {
	var c Cleanup
	var order []int
	for i := 1; i <= 3; i++ {
		c.Add(func() error { order = append(order, i); return nil })
	}
	if err := c.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !reflect.DeepEqual(order, []int{3, 2, 1}) {
		t.Fatalf("порядок %v, ожидали [3 2 1]", order)
	}
}

func TestCleanupCollectsErrors(t *testing.T) {
	var c Cleanup
	errA, errB := errors.New("a"), errors.New("b")
	ran := 0
	c.Add(func() error { ran++; return errA })
	c.Add(func() error { ran++; return nil })
	c.Add(func() error { ran++; return errB })
	err := c.Run()
	if ran != 3 {
		t.Fatalf("выполнено %d функций из 3: ошибка одной не должна останавливать остальные", ran)
	}
	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Fatalf("в ошибке должны быть обе: %v", err)
	}
}

func TestCleanupSurvivesPanic(t *testing.T) {
	var c Cleanup
	first := false
	c.Add(func() error { first = true; return nil })
	c.Add(func() error { panic("сломалось") })
	err := c.Run()
	if !first {
		t.Fatal("после паники в одной очистке остальные тоже должны выполниться")
	}
	if err == nil || !strings.Contains(err.Error(), "сломалось") {
		t.Fatalf("паника должна стать ошибкой с её текстом, получили %v", err)
	}
}

func TestCleanupRunTwice(t *testing.T) {
	var c Cleanup
	n := 0
	c.Add(func() error { n++; return nil })
	c.Run()
	if err := c.Run(); err != nil || n != 1 {
		t.Fatalf("повторный Run не должен ничего делать: n=%d err=%v", n, err)
	}
}
