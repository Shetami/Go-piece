package main

var chkBoom = errors.New("disk full")

type chkBadWriter struct{ calls int }

func (w *chkBadWriter) Write(p []byte) (int, error) { w.calls++; return 0, chkBoom }

type chkShortWriter struct{ calls int }

func (w *chkShortWriter) Write(p []byte) (int, error) { w.calls++; return len(p) / 2, nil }

func TestFanoutAll(t *testing.T) {
	var a, b bytes.Buffer
	var f io.Writer = NewFanout(&a, &b)
	fmt.Fprintf(f, "id=%d;", 7)
	f.Write([]byte("ok"))
	if a.String() != "id=7;ok" || b.String() != "id=7;ok" {
		t.Fatalf("получили %q и %q, ожидали \"id=7;ok\" в обоих", a.String(), b.String())
	}
}

func TestFanoutDropsFailed(t *testing.T) {
	var a, c bytes.Buffer
	bad := &chkBadWriter{}
	f := NewFanout(&a, bad, &c)
	n, err := f.Write([]byte("abc"))
	if n != 3 || !errors.Is(err, chkBoom) {
		t.Fatalf("Write = %d, %v; ожидали 3 и ошибку, для которой errors.Is(err, chkBoom)", n, err)
	}
	if !strings.Contains(err.Error(), "writer 1:") {
		t.Fatalf("текст ошибки %q должен содержать \"writer 1:\"", err)
	}
	n, err = f.Write([]byte("de"))
	if n != 2 || err != nil {
		t.Fatalf("второй Write = %d, %v; упавший writer должен выбыть, ожидали 2, nil", n, err)
	}
	if bad.calls != 1 || f.Alive() != 2 {
		t.Fatalf("упавшего вызвали %d раз, Alive = %d; ожидали 1 и 2", bad.calls, f.Alive())
	}
	if c.String() != "abcde" {
		t.Fatalf("writer после упавшего получил %q, ожидали \"abcde\"", c.String())
	}
}

func TestFanoutShortWrite(t *testing.T) {
	var a bytes.Buffer
	s := &chkShortWriter{}
	f := NewFanout(s, &a)
	_, err := f.Write([]byte("abcd"))
	if !errors.Is(err, io.ErrShortWrite) || !strings.Contains(err.Error(), "writer 0:") {
		t.Fatalf("короткая запись без ошибки: получили %v, ожидали io.ErrShortWrite с \"writer 0:\"", err)
	}
	f.Write([]byte("x"))
	if s.calls != 1 || f.Alive() != 1 {
		t.Fatalf("недописавший writer должен выбыть: вызовов %d, Alive %d", s.calls, f.Alive())
	}
}

func TestFanoutAllFail(t *testing.T) {
	f := NewFanout(nil, &chkBadWriter{}, &chkShortWriter{})
	n, err := f.Write([]byte("zz"))
	if n != 0 || !errors.Is(err, chkBoom) || !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("все упали: Write = %d, %v; ожидали 0 и обе ошибки", n, err)
	}
	if !strings.Contains(err.Error(), "writer 1:") || !strings.Contains(err.Error(), "writer 2:") {
		t.Fatalf("индексы считаются по исходному списку (nil тоже занимает место): %q", err)
	}
	n, err = f.Write([]byte("zz"))
	if n != 0 || !errors.Is(err, ErrNoWriters) {
		t.Fatalf("после выбывания всех: %d, %v; ожидали 0, ErrNoWriters", n, err)
	}
	if n, err := NewFanout().Write([]byte("a")); n != 0 || err != ErrNoWriters {
		t.Fatalf("пустой Fanout: %d, %v; ожидали 0, ErrNoWriters", n, err)
	}
}
