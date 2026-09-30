package main

type chkSrc struct {
	r        io.Reader
	closeErr error
	log      *[]string
}

func (s *chkSrc) Read(p []byte) (int, error) { return s.r.Read(p) }
func (s *chkSrc) Close() error              { *s.log = append(*s.log, "src"); return s.closeErr }

type chkDst struct {
	buf      bytes.Buffer
	limit    int // сколько байт примет до ошибки; <0 — без лимита
	panicky  bool
	closeErr error
	log      *[]string
}

var chkErrDiskFull = errors.New("диск заполнен")

func (d *chkDst) Write(p []byte) (int, error) {
	if d.panicky {
		panic("сломанный writer")
	}
	if d.limit >= 0 && d.buf.Len()+len(p) > d.limit {
		k := d.limit - d.buf.Len()
		d.buf.Write(p[:k])
		return k, chkErrDiskFull
	}
	return d.buf.Write(p)
}
func (d *chkDst) Close() error { *d.log = append(*d.log, "dst"); return d.closeErr }

func TestCopyAndCloseOK(t *testing.T) {
	var log []string
	dst := &chkDst{limit: -1, log: &log}
	n, err := CopyAndClose(dst, &chkSrc{r: strings.NewReader("hello"), log: &log})
	if err != nil || n != 5 || dst.buf.String() != "hello" {
		t.Fatalf("n=%d err=%v данные=%q, ожидали 5, nil, \"hello\"", n, err, dst.buf.String())
	}
	if !reflect.DeepEqual(log, []string{"dst", "src"}) {
		t.Fatalf("порядок закрытия %v, ожидали [dst src]", log)
	}
}

func TestCopyAndCloseDstCloseError(t *testing.T) {
	var log []string
	flushErr := errors.New("flush не удался")
	n, err := CopyAndClose(&chkDst{limit: -1, closeErr: flushErr, log: &log},
		&chkSrc{r: strings.NewReader("abc"), log: &log})
	if !errors.Is(err, flushErr) || !strings.Contains(err.Error(), "close dst") {
		t.Fatalf("err = %v, ожидали ошибку \"close dst: ...\" — ошибку Close писателя терять нельзя", err)
	}
	if n != 3 {
		t.Fatalf("n = %d, ожидали 3: ошибка закрытия не отменяет скопированное", n)
	}
}

func TestCopyAndCloseAllErrors(t *testing.T) {
	var log []string
	errD, errS := errors.New("dst"), errors.New("src")
	n, err := CopyAndClose(&chkDst{limit: 4, closeErr: errD, log: &log},
		&chkSrc{r: strings.NewReader("0123456789"), closeErr: errS, log: &log})
	if n != 4 {
		t.Fatalf("n = %d, ожидали 4 байта до ошибки", n)
	}
	for _, want := range []error{chkErrDiskFull, errD, errS} {
		if !errors.Is(err, want) {
			t.Fatalf("err = %v, не нашли %v — нужны все ошибки", err, want)
		}
	}
	for _, s := range []string{"copy:", "close dst:", "close src:"} {
		if !strings.Contains(err.Error(), s) {
			t.Fatalf("err = %q, нет подписи %q", err, s)
		}
	}
	if !reflect.DeepEqual(log, []string{"dst", "src"}) {
		t.Fatalf("после ошибки копирования закрыто %v, ожидали [dst src]", log)
	}
}

func TestCopyAndClosePanic(t *testing.T) {
	var log []string
	defer func() {
		if r := recover(); r != "сломанный writer" {
			t.Fatalf("паника должна пролететь дальше, recover() = %v", r)
		}
		if !reflect.DeepEqual(log, []string{"dst", "src"}) {
			t.Fatalf("при панике закрыто %v, ожидали [dst src]", log)
		}
	}()
	CopyAndClose(&chkDst{panicky: true, log: &log}, &chkSrc{r: strings.NewReader("x"), log: &log})
	t.Fatal("паника проглочена")
}
