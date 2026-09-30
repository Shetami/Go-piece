package main

// chkSteps — Read по сценарию: каждый шаг отдаёт строку и ошибку.
type chkStep struct {
	s   string
	err error
}

type chkSteps struct{ steps []chkStep }

func (r *chkSteps) Read(p []byte) (int, error) {
	if len(r.steps) == 0 {
		return 0, io.EOF
	}
	st := r.steps[0]
	r.steps = r.steps[1:]
	return copy(p, st.s), st.err
}

// chkPlainW — только Write, без ReadFrom (bytes.Buffer его умеет, поэтому прячем).
type chkPlainW struct{ buf bytes.Buffer }

func (w *chkPlainW) Write(p []byte) (int, error) { return w.buf.Write(p) }

type chkFastSrc struct{ used bool }

func (s *chkFastSrc) Read(p []byte) (int, error) { panic("есть WriteTo — Read звать не нужно") }
func (s *chkFastSrc) WriteTo(w io.Writer) (int64, error) {
	s.used = true
	n, err := io.WriteString(w, "fast")
	return int64(n), err
}

type chkRFDst struct {
	chkPlainW
	used bool
}

func (d *chkRFDst) ReadFrom(r io.Reader) (int64, error) {
	d.used = true
	b, err := io.ReadAll(r)
	d.buf.Write(b)
	return int64(len(b)), err
}

type chkLimitW struct{ max int }

func (w *chkLimitW) Write(p []byte) (int, error) { return min(len(p), w.max), nil }

type chkLiarW struct{}

func (chkLiarW) Write(p []byte) (int, error) { return len(p) + 5, nil }

func TestCopyFastPaths(t *testing.T) {
	src := &chkFastSrc{}
	dst := &chkRFDst{}
	n, err := Copy(dst, src)
	if !src.used || dst.used || n != 4 || err != nil || dst.buf.String() != "fast" {
		t.Fatalf("src с WriteTo и dst с ReadFrom: WriteTo=%v ReadFrom=%v n=%d err=%v; WriteTo должен иметь приоритет", src.used, dst.used, n, err)
	}
	dst = &chkRFDst{}
	n, err = Copy(dst, &chkSteps{steps: []chkStep{{"ab", nil}, {"c", nil}}})
	if !dst.used || n != 3 || err != nil || dst.buf.String() != "abc" {
		t.Fatalf("dst с ReadFrom: ReadFrom=%v n=%d err=%v %q; ожидали вызов ReadFrom", dst.used, n, err, dst.buf.String())
	}
}

func TestCopyLoop(t *testing.T) {
	var w chkPlainW
	src := &chkSteps{steps: []chkStep{{"he", nil}, {"", nil}, {"", nil}, {"llo", io.EOF}}}
	n, err := Copy(&w, src)
	if n != 5 || err != nil || w.buf.String() != "hello" {
		t.Fatalf("Copy = %d, %v, %q; ожидали 5, nil, \"hello\" — данные вместе с EOF тоже пишутся, 0, nil — не конец", n, err, w.buf.String())
	}
	// strings.Reader умеет WriteTo — прячем его, чтобы проверить цикл на большом объёме.
	big := strings.Repeat("0123456789", 10_000)
	w = chkPlainW{}
	n, err = Copy(&w, struct{ io.Reader }{strings.NewReader(big)})
	if n != int64(len(big)) || err != nil || w.buf.String() != big {
		t.Fatalf("100 КБ через цикл: n=%d err=%v; ожидали %d байт без ошибки", n, err, len(big))
	}
}

func TestCopyReadError(t *testing.T) {
	boom := errors.New("conn reset")
	var w chkPlainW
	n, err := Copy(&w, &chkSteps{steps: []chkStep{{"ab", nil}, {"cd", boom}}})
	if n != 4 || err != boom || w.buf.String() != "abcd" {
		t.Fatalf("ошибка чтения вместе с данными: n=%d err=%v %q; ожидали 4, conn reset, \"abcd\"", n, err, w.buf.String())
	}
}

func TestCopyBadWriters(t *testing.T) {
	n, err := Copy(&chkLimitW{max: 3}, struct{ io.Reader }{strings.NewReader("abcdef")})
	if n != 3 || err != io.ErrShortWrite {
		t.Fatalf("writer пишет не больше 3 байт без ошибки: n=%d err=%v; ожидали 3, io.ErrShortWrite", n, err)
	}
	n, err = Copy(chkLiarW{}, struct{ io.Reader }{strings.NewReader("abc")})
	if n != 0 || err != ErrInvalidWrite {
		t.Fatalf("writer вернул n > len(p): n=%d err=%v; ожидали 0, ErrInvalidWrite", n, err)
	}
}
