package main

type chkSource struct {
	lines    []string
	failAt   int // на каком Next вернуть ошибку; 0 — никогда
	nexts    int
	closes   int
	closeErr error
}

var chkErrRead = errors.New("соединение сброшено")

func (s *chkSource) Next() (string, bool, error) {
	s.nexts++
	if s.nexts == s.failAt {
		return "", false, chkErrRead
	}
	if len(s.lines) == 0 {
		return "", false, nil
	}
	l := s.lines[0]
	s.lines = s.lines[1:]
	return l, true, nil
}

func (s *chkSource) Close() error { s.closes++; return s.closeErr }

func chkOpen(src *chkSource, opens *int) func() (Source, error) {
	return func() (Source, error) { *opens++; return src, nil }
}

func TestLinesFull(t *testing.T) {
	src := &chkSource{lines: []string{"a", "b", "c"}}
	opens := 0
	seq := Lines(chkOpen(src, &opens))
	if opens != 0 {
		t.Fatal("open вызван при вызове Lines — источник должен открываться, когда начинается range")
	}
	var got []string
	for l, err := range seq {
		if err != nil {
			t.Fatalf("неожиданная ошибка %v", err)
		}
		got = append(got, l)
	}
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) || src.closes != 1 {
		t.Fatalf("строки %v, Close вызван %d раз; ожидали [a b c] и 1", got, src.closes)
	}
}

func TestLinesBreak(t *testing.T) {
	src := &chkSource{lines: []string{"a", "b", "c", "d"}, closeErr: errors.New("close")}
	opens := 0
	n := 0
	for range Lines(chkOpen(src, &opens)) {
		n++
		if n == 2 {
			break
		}
	}
	if src.closes != 1 || src.nexts != 2 {
		t.Fatalf("после break: Close %d раз, Next %d раз; ожидали 1 и 2", src.closes, src.nexts)
	}
}

func TestLinesPanicInBody(t *testing.T) {
	src := &chkSource{lines: []string{"a", "b"}}
	opens := 0
	func() {
		defer func() {
			if r := recover(); r != "плохая строка" {
				t.Fatalf("паника тела цикла должна пролететь, recover() = %v", r)
			}
		}()
		for range Lines(chkOpen(src, &opens)) {
			panic("плохая строка")
		}
	}()
	if src.closes != 1 {
		t.Fatalf("после паники в теле цикла Close вызван %d раз, ожидали 1", src.closes)
	}
}

func TestLinesNextError(t *testing.T) {
	src := &chkSource{lines: []string{"a", "b", "c"}, failAt: 3}
	opens := 0
	var got []string
	var errs []error
	for l, err := range Lines(chkOpen(src, &opens)) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		got = append(got, l)
	}
	if !reflect.DeepEqual(got, []string{"a", "b"}) || len(errs) != 1 || !errors.Is(errs[0], chkErrRead) {
		t.Fatalf("строки %v, ошибки %v; ожидали [a b] и одну ошибку чтения", got, errs)
	}
	if src.closes != 1 {
		t.Fatalf("после ошибки Close вызван %d раз, ожидали 1", src.closes)
	}
}

func TestLinesOpenAndCloseErrors(t *testing.T) {
	bad := errors.New("нет доступа")
	var errs []error
	for _, err := range Lines(func() (Source, error) { return nil, bad }) {
		errs = append(errs, err)
	}
	if len(errs) != 1 || errs[0] != bad {
		t.Fatalf("ошибка open: получили %v, ожидали одну пару с ней", errs)
	}

	cerr := errors.New("flush не удался")
	src := &chkSource{lines: []string{"a"}, closeErr: cerr}
	opens := 0
	errs = nil
	for _, err := range Lines(chkOpen(src, &opens)) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 1 || errs[0] != cerr {
		t.Fatalf("ошибка Close в конце: получили %v, ожидали её последней парой", errs)
	}
}

func TestLinesReopen(t *testing.T) {
	opens := 0
	seq := Lines(func() (Source, error) { opens++; return &chkSource{lines: []string{"x"}}, nil })
	for range seq {
	}
	for range seq {
	}
	if opens != 2 {
		t.Fatalf("два range — open вызван %d раз, ожидали 2", opens)
	}
}
