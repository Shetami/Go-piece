package main

var chkErrIO = errors.New("диск полон")

func chkBoomNilMap() error {
	var m map[string]int
	m["x"] = 1
	return nil
}

func TestRunAllOK(t *testing.T) {
	var n atomic.Int64
	err := RunAll(func() error { n.Add(1); return nil }, func() error { n.Add(1); return nil })
	if err != nil || n.Load() != 2 {
		t.Fatalf("все задачи успешны: err=%#v выполнено=%d, ожидали nil и 2", err, n.Load())
	}
	if err := RunAll(); err != nil {
		t.Fatalf("без задач: ожидали nil, получили %v", err)
	}
}

func TestRunAllErrorsOrdered(t *testing.T) {
	err := RunAll(
		func() error { time.Sleep(30 * time.Millisecond); return chkErrIO },
		func() error { return nil },
		func() error { return io.EOF },
	)
	want := "задача 0: диск полон\nзадача 2: EOF"
	if err == nil || err.Error() != want {
		t.Fatalf("ошибка %q, ожидали %q (в порядке задач)", err, want)
	}
	if !errors.Is(err, chkErrIO) || !errors.Is(err, io.EOF) {
		t.Fatal("errors.Is должен находить исходные ошибки задач")
	}
}

func TestRunAllPanics(t *testing.T) {
	var after atomic.Bool
	err := RunAll(
		func() error { panic("всё сломалось") },
		chkBoomNilMap,
		func() error { time.Sleep(20 * time.Millisecond); after.Store(true); return nil },
		func() error { panic(fmt.Errorf("обёртка: %w", chkErrIO)) },
	)
	if !after.Load() {
		t.Fatal("паника в одной задаче помешала другим завершиться")
	}
	var pe *PanicError
	if !errors.As(err, &pe) || pe.Task != 0 || pe.Value != "всё сломалось" {
		t.Fatalf("ожидали первым *PanicError задачи 0 со значением паники, получили %v", err)
	}
	if !strings.Contains(err.Error(), "задача 0: паника: всё сломалось") {
		t.Fatalf("в тексте нет «задача 0: паника: всё сломалось»: %q", err)
	}
	var re runtime.Error
	if !errors.As(err, &re) {
		t.Fatal("паника рантайма (nil-мапа) должна доставаться через errors.As(err, &runtime.Error)")
	}
	if !errors.Is(err, chkErrIO) {
		t.Fatal("паника ошибкой: errors.Is должен находить исходную ошибку")
	}
	if pe.Unwrap() != nil {
		t.Fatal("паника строкой: Unwrap должен вернуть nil")
	}
}

func TestRunAllPanicStack(t *testing.T) {
	err := RunAll(func() error { return nil }, chkBoomNilMap)
	var pe *PanicError
	if !errors.As(err, &pe) || pe.Task != 1 {
		t.Fatalf("ожидали *PanicError задачи 1, получили %v", err)
	}
	if !strings.Contains(string(pe.Stack), "chkBoomNilMap") {
		t.Fatalf("в стеке нет функции, где случилась паника:\n%s", pe.Stack)
	}
}

func TestRunAllPanicNil(t *testing.T) {
	err := RunAll(func() error { panic(nil) })
	var pn *runtime.PanicNilError
	if !errors.As(err, &pn) {
		t.Fatalf("panic(nil) тоже надо поймать (*runtime.PanicNilError), получили %v", err)
	}
}
