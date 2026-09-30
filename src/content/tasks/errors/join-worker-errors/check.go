package main

var chkErrBad = errors.New("битая строка")

func TestProcessAllOK(t *testing.T) {
	var n atomic.Int64
	err := ProcessAll([]string{"a", "b", "c"}, 2, func(string) error { n.Add(1); return nil })
	if err != nil {
		t.Fatalf("все вызовы успешны, ожидали nil, получили %#v", err)
	}
	if n.Load() != 3 {
		t.Fatalf("f вызвана %d раз, ожидали 3", n.Load())
	}
	if err := ProcessAll(nil, 3, func(string) error { return chkErrBad }); err != nil {
		t.Fatalf("пустой вход: ожидали nil, получили %v", err)
	}
}

func TestProcessAllOrderByIndex(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	// Чем дальше элемент, тем быстрее он завершается: порядок завершения обратный.
	err := ProcessAll(items, 5, func(s string) error {
		time.Sleep(time.Duration('f'-s[0]) * 10 * time.Millisecond)
		if s == "b" || s == "d" {
			return fmt.Errorf("%s: %w", s, chkErrBad)
		}
		return nil
	})
	want := "item 1: b: битая строка\nitem 3: d: битая строка"
	if err == nil || err.Error() != want {
		t.Fatalf("текст ошибки = %q, ожидали %q (все неудачи, по порядку индексов)", err, want)
	}
	if !errors.Is(err, chkErrBad) {
		t.Fatalf("errors.Is не находит исходную ошибку в %v", err)
	}
}

func TestProcessAllAs(t *testing.T) {
	err := ProcessAll([]string{"x", "y"}, 1, func(s string) error {
		if s == "y" {
			return &strconv.NumError{Func: "Atoi", Num: s, Err: strconv.ErrSyntax}
		}
		return nil
	})
	var ne *strconv.NumError
	if !errors.As(err, &ne) || ne.Num != "y" {
		t.Fatalf("errors.As должен достать *strconv.NumError для \"y\", ошибка: %v", err)
	}
}

func TestProcessAllLimit(t *testing.T) {
	var cur, peak atomic.Int64
	items := make([]string, 20)
	err := ProcessAll(items, 3, func(string) error {
		n := cur.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		cur.Add(-1)
		return chkErrBad
	})
	if peak.Load() > 3 {
		t.Fatalf("одновременно работало %d вызовов, лимит 3", peak.Load())
	}
	if peak.Load() < 2 {
		t.Fatalf("пик параллельности %d — вызовы идут последовательно", peak.Load())
	}
	if err == nil {
		t.Fatal("все 20 вызовов упали, а ошибка nil")
	}
	if got := strings.Count(err.Error(), "\n") + 1; got != 20 {
		t.Fatalf("в ошибке %d неудач, ожидали все 20", got)
	}
}

func TestProcessAllNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 20 {
		ProcessAll([]string{"a", "b", "c", "d"}, 8, func(string) error { return chkErrBad })
	}
	time.Sleep(50 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("горутин было %d, стало %d — воркеры не завершаются", before, after)
	}
}
