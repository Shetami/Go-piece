package main

// chkPart — часть с журналом; last — ошибка вместе с последним куском.
type chkPart struct {
	name     string
	chunks   []string
	last     error
	closeErr error
	log      *[]string
}

func (p *chkPart) Read(b []byte) (int, error) {
	if len(p.chunks) == 0 {
		if p.last != nil {
			return 0, p.last
		}
		return 0, io.EOF
	}
	n := copy(b, p.chunks[0])
	p.chunks = p.chunks[1:]
	if len(p.chunks) == 0 && p.last == io.EOF {
		return n, io.EOF
	}
	return n, nil
}

func (p *chkPart) Close() error {
	*p.log = append(*p.log, "close "+p.name)
	return p.closeErr
}

func chkReadAll(r io.Reader) (string, []error) {
	var out []byte
	var errs []error
	buf := make([]byte, 16)
	for range 100 {
		n, err := r.Read(buf)
		out = append(out, buf[:n]...)
		if err == io.EOF {
			return string(out), errs
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	return string(out), append(errs, errors.New("нет io.EOF за 100 вызовов"))
}

func TestConcatAll(t *testing.T) {
	var log []string
	c := Concat(
		&chkPart{name: "a", chunks: []string{"he", "l"}, log: &log},
		&chkPart{name: "b", chunks: []string{"lo"}, last: io.EOF, log: &log},
		&chkPart{name: "empty", log: &log},
		&chkPart{name: "c", chunks: []string{"!"}, last: io.EOF, log: &log},
	)
	got, errs := chkReadAll(c)
	if got != "hello!" || len(errs) != 0 {
		t.Fatalf("прочитали %q, ошибки %v; ожидали \"hello!\" без ошибок", got, errs)
	}
	if strings.Join(log, "; ") != "close a; close b; close empty; close c" {
		t.Fatalf("журнал закрытий %q; каждая часть закрывается сразу после своего EOF", log)
	}
	if err := c.Close(); err != nil || len(log) != 4 {
		t.Fatalf("Close после полного чтения: %v, журнал %q; повторно закрывать нельзя", err, log)
	}
	if got, errs := chkReadAll(Concat()); got != "" || len(errs) != 0 {
		t.Fatalf("Concat() без частей: %q, %v", got, errs)
	}
}

func TestConcatEarlyClose(t *testing.T) {
	var log []string
	c := Concat(
		&chkPart{name: "a", chunks: []string{"x"}, log: &log},
		&chkPart{name: "b", chunks: []string{"y", "z"}, log: &log, closeErr: errors.New("b failed")},
		&chkPart{name: "c", chunks: []string{"w"}, log: &log, closeErr: errors.New("c failed")},
	)
	buf := make([]byte, 1)
	c.Read(buf) // x
	c.Read(buf) // EOF a → close a, дальше b
	c.Read(buf)
	err := c.Close()
	if strings.Join(log, "; ") != "close a; close b; close c" {
		t.Fatalf("журнал %q; Close должен закрыть недочитанную и неначатую части по порядку", log)
	}
	if err == nil || !strings.Contains(err.Error(), "b failed") || !strings.Contains(err.Error(), "c failed") {
		t.Fatalf("Close = %v; ожидали обе ошибки через errors.Join", err)
	}
	if err := c.Close(); err != nil || len(log) != 3 {
		t.Fatalf("повторный Close = %v, журнал %q", err, log)
	}
	if n, err := c.Read(buf); n != 0 || err != ErrClosed {
		t.Fatalf("Read после Close = %d, %v; ожидали 0, ErrClosed", n, err)
	}
}

func TestConcatCloseErrorOnEOF(t *testing.T) {
	var log []string
	fail := errors.New("disk")
	c := Concat(
		&chkPart{name: "a", chunks: []string{"ab"}, last: io.EOF, log: &log, closeErr: fail},
		&chkPart{name: "b", chunks: []string{"cd"}, log: &log},
	)
	n, err := c.Read(make([]byte, 8))
	if n != 2 || !errors.Is(err, fail) || !strings.Contains(err.Error(), "part 0") {
		t.Fatalf("ошибка закрытия исчерпанной части: %d, %v; ожидали 2 и \"part 0: disk\"", n, err)
	}
	got, errs := chkReadAll(c)
	if got != "cd" || len(errs) != 0 {
		t.Fatalf("после ошибки закрытия чтение продолжается: %q, %v", got, errs)
	}
	c.Close()
	if strings.Join(log, "; ") != "close a; close b" {
		t.Fatalf("журнал %q; часть a уже закрыта — второй раз нельзя", log)
	}
}

func TestConcatReadError(t *testing.T) {
	var log []string
	boom := errors.New("conn reset")
	c := Concat(&chkPart{name: "a", chunks: []string{"ab"}, last: boom, log: &log}, &chkPart{name: "b", log: &log})
	buf := make([]byte, 8)
	n, _ := c.Read(buf)
	got := string(buf[:n])
	_, err := c.Read(buf)
	if got != "ab" || err != boom || len(log) != 0 {
		t.Fatalf("ошибка части: %q, %v, журнал %q; ожидали ошибку как есть и без автозакрытия", got, err, log)
	}
	c.Close()
	if strings.Join(log, "; ") != "close a; close b" {
		t.Fatalf("Close после ошибки чтения: журнал %q", log)
	}
}
