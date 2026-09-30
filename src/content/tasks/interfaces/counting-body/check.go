package main

var chkCloseErr = errors.New("close failed")

// chkSrc отдаёт chunks по одному за Read; last — ошибка вместе с последним куском.
type chkSrc struct {
	chunks []string
	last   error
	reads  atomic.Int32
	closes atomic.Int32
}

func (s *chkSrc) Read(p []byte) (int, error) {
	s.reads.Add(1)
	if len(s.chunks) == 0 {
		return 0, io.EOF
	}
	n := copy(p, s.chunks[0])
	s.chunks = s.chunks[1:]
	if len(s.chunks) == 0 && s.last != nil {
		return n, s.last
	}
	return n, nil
}

func (s *chkSrc) Close() error {
	if s.closes.Add(1) == 1 {
		return chkCloseErr
	}
	return errors.New("закрыли повторно")
}

type chkDone struct {
	calls int
	n     int64
	err   error
}

func (d *chkDone) f(n int64, err error) { d.calls++; d.n, d.err = n, err }

func TestBodyReadAll(t *testing.T) {
	src := &chkSrc{chunks: []string{"hel", "lo"}}
	var d chkDone
	b := NewBody(src, d.f)
	got, err := io.ReadAll(b)
	if err != nil || string(got) != "hello" {
		t.Fatalf("ReadAll = %q, %v", got, err)
	}
	if d.calls != 1 || d.n != 5 || d.err != nil {
		t.Fatalf("после EOF onDone: вызовов %d, n=%d, err=%v; ожидали 1, 5, nil", d.calls, d.n, d.err)
	}
	if err := b.Close(); err != chkCloseErr {
		t.Fatalf("первый Close = %v, ожидали ошибку нижнего Close", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("повторный Close = %v, ожидали nil", err)
	}
	if d.calls != 1 || src.closes.Load() != 1 {
		t.Fatalf("onDone вызван %d раз, нижний Close — %d; ожидали по 1", d.calls, src.closes.Load())
	}
}

func TestBodyDataWithEOF(t *testing.T) {
	src := &chkSrc{chunks: []string{"abc"}, last: io.EOF}
	var d chkDone
	b := NewBody(src, d.f)
	n, err := b.Read(make([]byte, 10))
	if n != 3 || err != io.EOF || d.calls != 1 || d.n != 3 {
		t.Fatalf("данные вместе с EOF: Read = %d, %v; onDone(%d) вызван %d раз; ожидали n=3 в onDone", n, err, d.n, d.calls)
	}
}

func TestBodyReadError(t *testing.T) {
	boom := errors.New("conn reset")
	src := &chkSrc{chunks: []string{"ab", "c"}, last: boom}
	var d chkDone
	b := NewBody(src, d.f)
	io.ReadAll(b)
	if d.calls != 1 || d.n != 3 || d.err != boom {
		t.Fatalf("ошибка чтения: onDone(%d, %v) вызван %d раз; ожидали (3, conn reset) один раз", d.n, d.err, d.calls)
	}
	b.Close()
	if d.calls != 1 || src.closes.Load() != 1 {
		t.Fatalf("Close после ошибки: onDone вызван %d раз, нижний Close — %d; ожидали 1 и 1", d.calls, src.closes.Load())
	}
}

func TestBodyEarlyClose(t *testing.T) {
	src := &chkSrc{chunks: []string{"ab", "cd", "ef"}}
	var d chkDone
	b := NewBody(src, nil) // nil onDone не должен ронять
	b.Read(make([]byte, 8))
	b.Close()
	src2 := &chkSrc{chunks: []string{"ab", "cd", "ef"}}
	b = NewBody(src2, d.f)
	b.Read(make([]byte, 8))
	b.Close()
	if d.calls != 1 || d.n != 2 || d.err != nil {
		t.Fatalf("Close на середине: onDone(%d, %v) вызван %d раз; ожидали (2, nil) один раз", d.n, d.err, d.calls)
	}
	reads := src2.reads.Load()
	n, err := b.Read(make([]byte, 8))
	if n != 0 || err != ErrBodyClosed || src2.reads.Load() != reads {
		t.Fatalf("Read после Close = %d, %v; ожидали 0, ErrBodyClosed без обращения к источнику", n, err)
	}
}

func TestBodyConcurrentClose(t *testing.T) {
	src := &chkSrc{chunks: []string{"x"}}
	var calls atomic.Int32
	b := NewBody(src, func(int64, error) { calls.Add(1) })
	var wg sync.WaitGroup
	var firstErrs atomic.Int32
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Close() != nil {
				firstErrs.Add(1)
			}
		}()
	}
	wg.Wait()
	if src.closes.Load() != 1 || calls.Load() != 1 || firstErrs.Load() != 1 {
		t.Fatalf("20 параллельных Close: нижний Close %d, onDone %d, ошибку получили %d; ожидали 1, 1, 1",
			src.closes.Load(), calls.Load(), firstErrs.Load())
	}
}
