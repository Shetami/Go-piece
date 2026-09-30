package main

// chkBase — нижний writer: журнал заголовков и тело.
type chkBase struct {
	headers []int
	body    bytes.Buffer
	flushes int
	short   bool // писать на 1 байт меньше
}

func (b *chkBase) WriteHeader(code int) { b.headers = append(b.headers, code) }
func (b *chkBase) Write(p []byte) (int, error) {
	if b.short && len(p) > 0 {
		b.body.Write(p[:len(p)-1])
		return len(p) - 1, io.ErrShortWrite
	}
	return b.body.Write(p)
}

type chkF struct{ *chkBase }

func (f chkF) Flush() { f.flushes++ }

type chkRF struct{ *chkBase }

func (r chkRF) ReadFrom(src io.Reader) (int64, error) { return r.body.ReadFrom(src) }

type chkFRF struct{ *chkBase }

func (x chkFRF) Flush()                                { x.flushes++ }
func (x chkFRF) ReadFrom(src io.Reader) (int64, error) { return x.body.ReadFrom(src) }

func TestRecorderStatus(t *testing.T) {
	base := &chkBase{}
	var rec Recorder
	w := Wrap(base, &rec)
	w.WriteHeader(404)
	w.WriteHeader(500)
	fmt.Fprint(w, "not found")
	if rec.Status != 404 || fmt.Sprint(base.headers) != "[404]" || rec.Bytes != 9 {
		t.Fatalf("Status=%d, в w ушли заголовки %v, Bytes=%d; ожидали 404, [404], 9", rec.Status, base.headers, rec.Bytes)
	}
	base = &chkBase{}
	rec = Recorder{}
	w = Wrap(base, &rec)
	w.Write([]byte("ok"))
	w.WriteHeader(503)
	if rec.Status != 200 || len(base.headers) != 0 {
		t.Fatalf("Write до WriteHeader: Status=%d, заголовки в w %v; ожидали 200 и поздний WriteHeader проигнорирован", rec.Status, base.headers)
	}
}

func TestRecorderShortWrite(t *testing.T) {
	var rec Recorder
	n, err := Wrap(&chkBase{short: true}, &rec).Write([]byte("hello"))
	if n != 4 || err == nil || rec.Bytes != 4 {
		t.Fatalf("частичная запись: n=%d err=%v Bytes=%d; ожидали 4, ошибку и Bytes=4", n, err, rec.Bytes)
	}
}

func TestRecorderOptional(t *testing.T) {
	cases := []struct {
		name      string
		w         ResponseWriter
		flush, rf bool
	}{
		{"plain", &chkBase{}, false, false},
		{"flusher", chkF{&chkBase{}}, true, false},
		{"readerfrom", chkRF{&chkBase{}}, false, true},
		{"both", chkFRF{&chkBase{}}, true, true},
	}
	for _, c := range cases {
		var rec Recorder
		w := Wrap(c.w, &rec)
		_, isF := w.(Flusher)
		_, isRF := w.(io.ReaderFrom)
		if isF != c.flush || isRF != c.rf {
			t.Fatalf("%s: обёртка Flusher=%v ReaderFrom=%v; ожидали %v и %v — ровно как у исходного", c.name, isF, isRF, c.flush, c.rf)
		}
		u, ok := w.(interface{ Unwrap() ResponseWriter })
		if !ok || u.Unwrap() != c.w {
			t.Fatalf("%s: Unwrap должен вернуть исходный writer", c.name)
		}
	}
}

func TestRecorderFlushAndReadFrom(t *testing.T) {
	base := &chkBase{}
	var rec Recorder
	w := Wrap(chkF{base}, &rec)
	w.(Flusher).Flush()
	w.WriteHeader(201)
	if base.flushes != 1 || rec.Status != 200 || len(base.headers) != 0 {
		t.Fatalf("Flush до WriteHeader: flushes=%d Status=%d заголовки %v; ожидали 1, 200 и поздний WriteHeader проигнорирован", base.flushes, rec.Status, base.headers)
	}
	base = &chkBase{}
	rec = Recorder{}
	w = Wrap(chkFRF{base}, &rec)
	n, err := w.(io.ReaderFrom).ReadFrom(strings.NewReader("streamed body"))
	w.Write([]byte("!"))
	if n != 13 || err != nil || rec.Bytes != 14 || rec.Status != 200 || base.body.String() != "streamed body!" {
		t.Fatalf("ReadFrom: n=%d err=%v Bytes=%d Status=%d тело %q", n, err, rec.Bytes, rec.Status, base.body.String())
	}
}
