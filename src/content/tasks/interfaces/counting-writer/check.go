package main

var _ io.Writer = (*CountingWriter)(nil)

func TestCountingWriterPassesThrough(t *testing.T) {
	var buf bytes.Buffer
	cw := &CountingWriter{W: &buf}
	fmt.Fprintf(cw, "привет, %s!", "мир")
	if buf.String() != "привет, мир!" {
		t.Fatalf("в W дошло %q, ожидали %q", buf.String(), "привет, мир!")
	}
	if cw.N != int64(len("привет, мир!")) {
		t.Fatalf("N = %d, ожидали %d (байты, а не символы)", cw.N, len("привет, мир!"))
	}
}

func TestCountingWriterAccumulates(t *testing.T) {
	cw := &CountingWriter{W: io.Discard}
	io.WriteString(cw, "abc")
	io.WriteString(cw, "de")
	if cw.N != 5 {
		t.Fatalf("после двух записей N = %d, ожидали 5", cw.N)
	}
}

type checkHalfWriter struct{}

func (checkHalfWriter) Write(p []byte) (int, error) {
	return len(p) / 2, errors.New("диск заполнен")
}

func TestCountingWriterPartialWrite(t *testing.T) {
	cw := &CountingWriter{W: checkHalfWriter{}}
	n, err := cw.Write([]byte("12345678"))
	if err == nil || err.Error() != "диск заполнен" {
		t.Fatalf("ошибка нижнего writer'а должна дойти как есть, получили %v", err)
	}
	if n != 4 || cw.N != 4 {
		t.Fatalf("записано 4 байта из 8: n=%d, N=%d", n, cw.N)
	}
}
