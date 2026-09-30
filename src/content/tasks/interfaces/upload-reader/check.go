package main

// chkRead читает r буфером размера buf до ошибки.
func chkRead(r io.Reader, buf int) (string, error) {
	var out []byte
	p := make([]byte, buf)
	for i := 0; i < 1000; i++ {
		n, err := r.Read(p)
		out = append(out, p[:n]...)
		if err != nil {
			return string(out), err
		}
	}
	return string(out), errors.New("Read ни разу не вернул ошибку")
}

func TestUploadFits(t *testing.T) {
	var sink bytes.Buffer
	got, err := chkRead(Upload(strings.NewReader("hello"), 5, &sink), 2)
	if got != "hello" || err != io.EOF || sink.String() != "hello" {
		t.Fatalf("ровно max байт: %q, %v, sink %q; ожидали hello, io.EOF, hello", got, err, sink.String())
	}
	sink.Reset()
	got, err = chkRead(Upload(strings.NewReader("hi"), 100, &sink), 64)
	if got != "hi" || err != io.EOF || sink.String() != "hi" {
		t.Fatalf("короче лимита: %q, %v, sink %q", got, err, sink.String())
	}
}

func TestUploadTooLarge(t *testing.T) {
	for _, buf := range []int{1, 3, 64} {
		var sink bytes.Buffer
		src := strings.NewReader("abcdefghij")
		u := Upload(src, 4, &sink)
		got, err := chkRead(u, buf)
		if got != "abcd" || err != ErrTooLarge || sink.String() != "abcd" {
			t.Fatalf("буфер %d, 10 байт при max 4: %q, %v, sink %q; ожидали abcd, ErrTooLarge, abcd", buf, got, err, sink.String())
		}
		if n, err := u.Read(make([]byte, 8)); n != 0 || err != ErrTooLarge {
			t.Fatalf("повторный Read после ErrTooLarge: %d, %v", n, err)
		}
		if src.Len() != 5 {
			t.Fatalf("из источника прочитано %d байт, ожидали не больше max+1 = 5", 10-src.Len())
		}
	}
}

func TestUploadZero(t *testing.T) {
	var sink bytes.Buffer
	if _, err := chkRead(Upload(strings.NewReader(""), 0, &sink), 8); err != io.EOF {
		t.Fatalf("max 0, пустой источник: %v, ожидали io.EOF", err)
	}
	if got, err := chkRead(Upload(strings.NewReader("x"), 0, &sink), 8); got != "" || err != ErrTooLarge || sink.Len() != 0 {
		t.Fatalf("max 0, один байт: %q, %v, sink %q; ожидали ErrTooLarge", got, err, sink.String())
	}
}

type chkFailSink struct{}

func (chkFailSink) Write(p []byte) (int, error) { return 0, errors.New("sink full") }

func TestUploadErrors(t *testing.T) {
	n, err := Upload(strings.NewReader("abc"), 10, chkFailSink{}).Read(make([]byte, 8))
	if n != 3 || err == nil || err.Error() != "sink full" {
		t.Fatalf("ошибка sink: %d, %v; ожидали 3 и sink full", n, err)
	}
	boom := errors.New("conn reset")
	var sink bytes.Buffer
	got, err := chkRead(Upload(io.MultiReader(strings.NewReader("ab"), iotest.ErrReader(boom)), 10, &sink), 8)
	if got != "ab" || err != boom {
		t.Fatalf("ошибка источника: %q, %v; ожидали ab и conn reset как есть", got, err)
	}
}
