package main

var chkErrParse = errors.New("не число")

func TestCollectOK(t *testing.T) {
	vals, err := Collect([]Result[int]{{Val: 1}, {Val: 2}, {Val: 3}})
	if err != nil || !slices.Equal(vals, []int{1, 2, 3}) {
		t.Fatalf("Collect = %v, %v; ожидали [1 2 3], nil", vals, err)
	}
	empty, err := Collect[string](nil)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("пустой вход: ожидали пустой не-nil слайс и nil, получили %#v, %v", empty, err)
	}
}

func TestCollectErrors(t *testing.T) {
	rs := []Result[int]{{Val: 1}, {Err: fmt.Errorf("строка «x»: %w", chkErrParse)}, {Val: 3}, {Err: io.EOF}}
	vals, err := Collect(rs)
	if vals != nil {
		t.Fatalf("при ошибках значения не возвращаются, а получили %v", vals)
	}
	want := "#1: строка «x»: не число\n#3: EOF"
	if err == nil || err.Error() != want {
		t.Fatalf("ошибка %q, ожидали %q", err, want)
	}
	if !errors.Is(err, chkErrParse) || !errors.Is(err, io.EOF) {
		t.Fatal("errors.Is должен находить каждую исходную ошибку")
	}
}

func TestMap(t *testing.T) {
	r := Map(Result[string]{Val: "42"}, strconv.Atoi)
	if r.Err != nil || r.Val != 42 {
		t.Fatalf("Map(\"42\", Atoi) = %+v, ожидали 42", r)
	}
	called := false
	r = Map(Result[string]{Err: chkErrParse}, func(string) (int, error) { called = true; return 0, nil })
	if called || r.Err != chkErrParse {
		t.Fatalf("при ошибке f не вызывается и ошибка переносится как есть: called=%v err=%v", called, r.Err)
	}
	var ne *strconv.NumError
	if r := Map(Result[string]{Val: "x"}, strconv.Atoi); !errors.As(r.Err, &ne) {
		t.Fatalf("ошибка f должна попасть в результат, получили %v", r.Err)
	}
}

func TestTry(t *testing.T) {
	r := Try(func() (int, error) { return 7, nil })
	if r.Val != 7 || r.Err != nil {
		t.Fatalf("Try без паники = %+v", r)
	}
	r = Try(func() (int, error) { panic("бум") })
	if !errors.Is(r.Err, ErrPanic) || !strings.Contains(fmt.Sprint(r.Err), "бум") {
		t.Fatalf("паника строкой: ожидали ошибку с ErrPanic и текстом «бум», получили %v", r.Err)
	}
	r = Try(func() (int, error) { panic(io.ErrUnexpectedEOF) })
	if !errors.Is(r.Err, ErrPanic) || !errors.Is(r.Err, io.ErrUnexpectedEOF) {
		t.Fatalf("паника ошибкой: errors.Is должен находить и ErrPanic, и исходную ошибку, получили %v", r.Err)
	}
}
