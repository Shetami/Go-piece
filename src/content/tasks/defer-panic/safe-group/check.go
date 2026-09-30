package main

type chkCodeErr struct{ code int }

func (e *chkCodeErr) Error() string { return "код " + strconv.Itoa(e.code) }

func chkExplode() error { panic("взрыв в задаче") }

func chkWaitDone(t *testing.T, g *Group) error {
	t.Helper()
	ch := make(chan error, 1)
	go func() { ch <- g.Wait() }()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("Wait не вернулся за 3 секунды")
		return nil
	}
}

func TestSafeGroupEmpty(t *testing.T) {
	var g Group
	if err := chkWaitDone(t, &g); err != nil {
		t.Fatalf("пустая группа: Wait() = %v, ожидали nil", err)
	}
}

func TestSafeGroupWaitsAll(t *testing.T) {
	var g Group
	var done atomic.Int32
	for i := 0; i < 5; i++ {
		g.Go(func() error {
			time.Sleep(time.Duration(i+1) * 5 * time.Millisecond)
			done.Add(1)
			return nil
		})
	}
	if err := chkWaitDone(t, &g); err != nil {
		t.Fatalf("Wait() = %v, ожидали nil", err)
	}
	if n := done.Load(); n != 5 {
		t.Fatalf("Wait вернулся, когда закончились %d задач из 5", n)
	}
}

func TestSafeGroupJoinsErrors(t *testing.T) {
	var g Group
	errA, errB := errors.New("a"), errors.New("b")
	g.Go(func() error { return errA })
	g.Go(func() error { return nil })
	g.Go(func() error { return errB })
	err := chkWaitDone(t, &g)
	if !errors.Is(err, errA) || !errors.Is(err, errB) {
		t.Fatalf("Wait() = %v, ожидали обе ошибки (errors.Is)", err)
	}
}

func TestSafeGroupPanicBecomesError(t *testing.T) {
	var g Group
	g.Go(chkExplode)
	g.Go(func() error { return nil })
	err := chkWaitDone(t, &g)
	var pe *PanicError
	if !errors.As(err, &pe) {
		t.Fatalf("Wait() = %v, ожидали *PanicError внутри", err)
	}
	if pe.Value != "взрыв в задаче" {
		t.Fatalf("PanicError.Value = %v, ожидали исходное значение паники", pe.Value)
	}
	if !strings.Contains(err.Error(), "взрыв в задаче") {
		t.Fatalf("текст ошибки %q не содержит значения паники", err.Error())
	}
	if !strings.Contains(string(pe.Stack), "chkExplode") {
		t.Fatalf("в Stack нет паникующей функции chkExplode — стек снят не там:\n%s", pe.Stack)
	}
}

func TestSafeGroupPanicWithError(t *testing.T) {
	var g Group
	g.Go(func() error { panic(&chkCodeErr{code: 42}) })
	g.Go(func() error {
		var m map[string]int
		m["x"] = 1
		return nil
	})
	err := chkWaitDone(t, &g)
	var ce *chkCodeErr
	if !errors.As(err, &ce) || ce.code != 42 {
		t.Fatalf("errors.As не нашёл исходную ошибку паники в %v", err)
	}
	var re runtime.Error
	if !errors.As(err, &re) {
		t.Fatalf("паника рантайма должна находиться через errors.As(err, *runtime.Error): %v", err)
	}
}
